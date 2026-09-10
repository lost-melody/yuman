# Yuman

> _Yuman_ is built for _Yume_, which is an _IME_ for Chinese,
> so I'll speak Chinese here.

_Yuman_ 是一個爲 _[Yume][yume-release]_ (宇夢輸入系統) 構建的 _CLI_ 工具,
僅支持 _Linux_ 操作系統 (`fcitx5-yume`).

[yume-release]: https://github.com/forfudan/yume-release

## Features

- 自動從 _GitHub_ 發佈頁安裝及更新 _Yume_: 執行 `yuman update-yume`.
- 從指定 _URL_ 或 `.tar.gz` 文件安裝 _Yume_: 執行 `yuman install [package]`.
- 重啓 _Fcitx5_ 服務或重載輸入法配置: `yuman restart` 及 `yuman reload [addon]`.
- 導入或移除自定義方案碼表: `yuman import [table.txt]` 及 `yuman yume custom remove <schema>`.
- 配置 _Yume_ 輸入系統: 執行 `yuman yume config` 打開 _TUI_ 配置工具.

## Development

主要依賴:

- [cobra][cobra]: 命令參數及子命令識别.
- [bubbletea][bubbletea]: 構建 _TUI_ 配置界面.
- [dbus][godbus]: 訪問 _Fcitx5_ 服務的 _DBus_ 接口.
- [i18n][go-i18n]: 實現文本多語言支持.

[cobra]: https://github.com/spf13/cobra
[bubbletea]: https://github.com/charmbracelet/bubbletea
[godbus]: https://github.com/godbus/dbus
[go-i18n]: https://github.com/nicksnyder/go-i18n

安裝與發佈:

- 直接安裝: `go install github.com/lost-melody/yuman@latest`.
- 手動編譯: 克隆並 `cd` 到倉庫目錄, 執行 `task build`.
- 打包 (含 `yuman` 文件和 `LICENSE`, `README.md`): 執行 `task pack`.

添加語言翻譯:

- 如須修訂已有翻譯文本, 可直接編輯相應文件: `nvim tr/locales/active.zh-CN.toml`.
- 若代碼中有新增的待翻譯字符串, 則先提取: `task i18n_extract`.
- 創建待翻譯稿件 (以 `zh-CN` 爲例): `task i18n_translate LANG=zh-CN`.
- 使用編輯器打開翻譯稿件 (以 `nvim` 和 `zh-CN` 爲例): `nvim tr/locales/translate.zh-CN.toml`.
- 翻譯完成後, 執行 `task i18n_apply LANG=zh-CN` 將新翻譯文本合併到已有翻譯文件.
