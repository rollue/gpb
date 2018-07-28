# Filesystem Backup


## GO ORM
- dependencies install
- archiver
- dirhash
- monitor
- backup
- backupd


### dependencies install
- go get github.com/stretchr/testify/require
- go get github.com/matryer/filedb


### archiver
- archive와 os모듈을 이용한 압축 모듈
- Archive
```go
w := zip.NewWriter(out)
defer w.Close()
return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
    if info.IsDir() {
        return nil // skip
    }
    if err != nil {
        return err
    }
    in, err := os.Open(path)
    if err != nil {
        return err
    }
    defer in.Close()
    f, err := w.Create(path)
    if err != nil {
        return err
    }
    _, err = io.Copy(f, in)
    if err != nil {
        return err
    }
    return nil
})
```
- Restore
```go
var w sync.WaitGroup
var errs []error
errChan := make(chan error)
go func() {
    for err := range errChan {
        errs = append(errs, err)
    }
}()
for _, f := range r.File {
    w.Add(1)
    go func(f *zip.File) {
        zippedfile, err := f.Open()
        if err != nil {
            errChan <- err
            w.Done()
            return
        }
        toFilename := path.Join(dest, f.Name)
        err = os.MkdirAll(path.Dir(toFilename), 0777)
        if err != nil {
            errChan <- err
            w.Done()
            return
        }
        newFile, err := os.Create(toFilename)
        if err != nil {
            zippedfile.Close()
            errChan <- err
            w.Done()
            return
        }
        _, err = io.Copy(newFile, zippedfile)
        newFile.Close()
        zippedfile.Close()
        if err != nil {
            errChan <- err
            w.Done()
            return
        }
        w.Done()
    }(f)
}
w.Wait()
```
- WaitGroup -> add시 delta 증가 / done시 delta 감소 
- delta값 0일시 블락된 고루틴 released


### dirhash
- 디렉토리의 해시값을 계산하는 모듈
```go
hash := md5.New()
err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
    if err != nil {
        return err
    }
    io.WriteString(hash, path)
    fmt.Fprintf(hash, "%v", info.IsDir())
    fmt.Fprintf(hash, "%v", info.ModTime())
    fmt.Fprintf(hash, "%v", info.Mode())
    fmt.Fprintf(hash, "%v", info.Name())
    fmt.Fprintf(hash, "%v", info.Size())
    return nil
})
if err != nil {
    return "", err
}
return fmt.Sprintf("%x", hash.Sum(nil)), nil
```

### monitor
- 디렉토리의 해시 값을 맵에 보관하여 비교
- 비교한 해시값이 다를 경우 archiver call
- now
```go
func (m *Monitor) Now() (int, error) {
	var counter int
	for path, lastHash := range m.Paths {
		newHash, err := DirHash(path)
		if err != nil {
			return 0, err
		}
		if newHash != lastHash {
			err := m.act(path)
			if err != nil {
				return counter, err
			}
			m.Paths[path] = newHash // update the hash
			counter++
		}
	}
	return counter, nil
}
```
- act
```go
func (m *Monitor) act(path string) error {
	dirname := filepath.Base(path)
	filename := fmt.Sprintf(m.Archiver.DestFmt(), time.Now().UnixNano())
	return m.Archiver.Archive(path, filepath.Join(m.Destination, dirname, filename))
}
```


### backup
- 백업할 디렉토리 리스트를 저장하는 모듈
- archiver를 이용한 실제 백업이 아닌 리스트를 백업!
- list
```go
var path path
col.ForEach(func(i int, data []byte) bool {
    err := json.Unmarshal(data, &path)
    if err != nil {
        fatalErr = err
        return false
    }
    fmt.Printf("= %s\n", path)
    return false
})
```
- add
```go
if len(args[1:]) == 0 {
    fatalErr = errors.New("must specify path to add")
    return
}
for _, p := range args[1:] {
    path := &path{Path: p, Hash: "Not yet archived"}
    if err := col.InsertJSON(path); err != nil {
        fatalErr = err
        return
    }
    fmt.Printf("+ %s\n", path)
}
```
- remove
```go
var path path
col.RemoveEach(func(i int, data []byte) (bool, bool) {
    err := json.Unmarshal(data, &path)
    if err != nil {
        fatalErr = err
        return false, true
    }
    for _, p := range args[1:] {
        if path.Path == p {
            fmt.Printf("- %s\n", path)
            return true, false
        }
    }
    return false, false
})
```


### backupd
- 실제 파일을 백업하는 모듈
- backup 모듈에서 저장된 디렉토리 리스트를 기반으로 monitoring 및 archive
- cache
```go
var path path
col.ForEach(func(_ int, data []byte) bool {
    if err := json.Unmarshal(data, &path); err != nil {
        fatalErr = err
        return true
    }
    m.Paths[path.Path] = path.Hash
    return false // carry on
})
```
- interval setting
```go
check(m, col)
signalChan := make(chan os.Signal, 1)
signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
for {
    select {
    case <-time.After(*interval):
        check(m, col)
    case <-signalChan:
        // stop
        fmt.Println()
        log.Printf("Stopping...")
        return
    }
}
```
- monitoring 및 backup
```go
func check(m *backup.Monitor, col *filedb.C) {
	log.Println("Checking...")
	counter, err := m.Now()
	if err != nil {
		log.Fatalln("failed to backup:", err)
	}
	if counter > 0 {
		log.Printf("  Archived %d directories\n", counter)
		// update hashes
		var path path
		col.SelectEach(func(_ int, data []byte) (bool, []byte, bool) {
			if err := json.Unmarshal(data, &path); err != nil {
				log.Println("failed to unmarshal data (skipping):", err)
				return true, data, false
			}
			path.Hash, _ = m.Paths[path.Path]
			newdata, err := json.Marshal(&path)
			if err != nil {
				log.Println("failed to marshal data (skipping):", err)
				return true, data, false
			}
			return true, newdata, false
		})
	} else {
		log.Println("  No changes")
	}
}
```



