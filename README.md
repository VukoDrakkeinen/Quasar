# Quasar Comic Reader

Why?
----

* Because why not?
* Because I wanted to try a new language (Go)
* Because I wanted to make something bigger than a Hello World

Features
--------

Not much yet. The project is extremely early in development.

Screenshots
-----------

Look above.

Roadmap
-------

- v.next:
  * [ ] rename? (for the third time!). Considered names: 
    1. Geon
    2. Gravastar
    3. Retrocausality
    4. Kugelblitz
    5. Lorentzian Manifold 

- v0.5:
  * [ ] ability to provide revenue to the hosting sites through shown ads (opt-in by default)
  * [ ] complete Batoto plugin
  
- v0.8:
  * [ ] more plugins
 
- v1.0:
  * [ ] seamless integration with numerous comic-reading sites 
  * [ ] scripted plugins for said sites


Quasar uses:
------------

1. Core: Go with (quite a lot!) of C/C++ glue code
2. GUI: [QtQuick](http://doc.qt.io/qt-5/qtquick-index.html) of [Qt 5](http://www.qt.io/) through [Go-QML](https://github.com/go-qml/qml) by Gustavo Niemeyer
3. Storage: [SQLite 3](https://www.sqlite.org/) through [go-sqlite3](https://github.com/mattn/go-sqlite3/) by mattn
4. Crypto: [libsodium](https://github.com/jedisct1/libsodium) through [libsodium-go](https://github.com/GoKillers/libsodium-go) by GoKillers
