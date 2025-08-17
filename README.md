# iOSDC Japan 2025 Albatross


# これは何？

2025-09-19 から 2025-09-21 にかけて開催された [iOSDC Japan 2025](https://iosdc.jp/2025/) の中の企画、Swift コードバトルのシステムです。

[サイトはこちら (現在は新規にプレイすることはできません)](https://t.nil.ninja/iosdc-japan/2025/code-battle/)


# サンドボックス化の仕組み

ユーザから任意のコードを受け付ける関係上、何も対策をしないと深刻な脆弱性を抱えてしまいます。

このシステムでは、送信された Swift コードを WebAssembly へとコンパイルして実行することで、サンドボックス化を実現しています。


# License

The contents of the repository are licensed under The MIT License, except for

* [backend/admin/assets/css/normalize.css](backend/admin/assets/normalize.css),
* [backend/admin/assets/css/sakura.css](backend/admin/assets/sakura.css),
* [frontend/public/favicon.svg](frontend/public/favicon.svg) and
* [frontend/public/logo.svg](frontend/public/logo.svg).

See [LICENSE](./LICENSE) for copylight notice.
