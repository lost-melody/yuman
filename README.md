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
