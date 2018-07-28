# GORM


## GO ORM

* [어떤것이 좋은가?](https://www.reddit.com/r/golang/comments/3ajqa6/golang_which_orm_is_better/)
* 가벼운 기능을 선호
* 모델만 잘 정의 되어 있다면 네이티브 쿼리나 ORM이나 상관없다!


## ORM

* gorm
 - struct를 기반으로 CRUD 기능을 제공한다. (기본적인 ORM 기능) 모델 간 Associations(belongs-to, has-one, has-many, many-to-many, polymorphism)를 정의할 수 있다. 하지만 실제 사용해보면 불편한 부분이 많다.
* xorm
 - struct를 기반으로 CRUD 기능을 제공한다(기본적인 ORM 기능). gorm과 유사하다. 모델 간 Associations을 정의하는 기능은 없다. 캐싱 기능을 제공한다. built-in 타입이 아닌 필드는 JSON으로 변환해 준다.


* GORM Quick Start
```go
package main

import (
    "github.com/jinzhu/gorm"
    _ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Product struct {
  gorm.Model
  Code string
  Price uint
}

func main() {
  db, err := gorm.Open("sqlite3", "test.db")
  if err != nil {
    panic("failed to connect database")
  }
  defer db.Close()

  // Migrate the schema
  db.AutoMigrate(&Product{})

  // Create
  db.Create(&Product{Code: "L1212", Price: 1000})

  // Read
  var product Product
  db.First(&product, 1) // find product with id 1
  db.First(&product, "code = ?", "L1212") // find product with code l1212

  // Update - update product's price to 2000
  db.Model(&product).Update("Price", 2000)

  // Delete - delete product
  db.Delete(&product)
}
```