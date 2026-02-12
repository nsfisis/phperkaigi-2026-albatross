# PHPerKaigi 2026 Albatross


# これは何？

2026-03-20 から 2026-03-22 にかけて開催された [PHPerKaigi 2026](https://phperkaigi.jp/2026/) の中の企画、PHPer コードバトルのシステムです。

[サイトはこちら (現在は新規にプレイすることはできません)](https://t.nil.ninja/phperkaigi/2026/code-battle/)


# サンドボックス化の仕組み

ユーザから任意のコードを受け付ける関係上、何も対策をしないと深刻な脆弱性を抱えてしまいます。

このシステムでは、送信されたコードを WebAssembly へ変換された PHP 処理系で実行することで、サンドボックス化を実現しています。


# License

The contents of the repository are licensed under The MIT License, except for

* [backend/admin/assets/css/normalize.css](backend/admin/assets/css/normalize.css),
* [backend/admin/assets/css/sakura.css](backend/admin/assets/css/sakura.css),
* [frontend/public/favicon.svg](frontend/public/favicon.svg) and
* [frontend/public/logo.svg](frontend/public/logo.svg).

See [LICENSE](./LICENSE) for copylight notice.
