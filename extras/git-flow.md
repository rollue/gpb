# Git-flow

## Reference

*  [우린 Git-flow를 사용하고 있어요 by 배달의민족](http://woowabros.github.io/experience/2017/10/30/baemin-mobile-git-branch-strategy.html)
* [Git-flow cheatsheet](https://danielkummer.github.io/git-flow-cheatsheet/index.ko_KR.html)



## 구조

* Upstream: 공유되는 최신 원격저장소
* Origin: forked upstream
* Local
* 참고: 저는 기존 project에서 origin 없이 바로 Upstream - Local로 사용했습니다.



## 브랜치

- master : 제품으로 출시될 수 있는 브랜치
- develop : 다음 출시 버전을 개발하는 브랜치
- feature : 기능을 개발하는 브랜치
- release : 이번 출시 버전을 준비하는 브랜치
- hotfix : 출시 버전에서 발생한 버그를 수정 하는 브랜치
- 참고: hotfix는 사용해본 적이 없습니다 ^^;;



## Git-flow

* 따라해보자
* jira와 연동법
  * ticket numbering!