#   Win Specs Reporter

A quick and portable tool for Windows inventory audit,
it comprises of two tiny tools:

1.  A launcher to show computer specs as a web page in the browser.

    The web page nicely fits an ISO A4 paper for printing,
    either as PDF or on plain paper.
    Always include background when printing!

2.  A CLI tool to dump the data as these formats,

    -   pretty-printed text (YAML-like),
    -   CSV,
    -   transposed CSV (headers in rows, instead of single row),
    -   JSON,
    -   YAML,
    -   TOML.

##  How to build

1.  Install Go and GNU Make,

    -   using WinGet,

        ```shell
        winget install GoLang.Go ezwinports.make
        ```

    -   or [scoop](https://scoop.sh/),

        ```shell
        scoop install go make
        ```

2.  Install `rsrc`,

    ```shell
    go get github.com/akavel/rsrc ; go install github.com/akavel/rsrc
    ```

3.  Run the Makefile,

    ```shell
    make
    ```

##  Editor config

The source uses custom build tags: `cli` and `gui`.
So, you need to configure your editor to recognize them.

>   [!TIP]
>   With Neovim and VS Code,
>   ignore `gopls` complaint about "main redeclared in this block"
>   in `cli.go` and `gui.go`.
>
>   GoLand stays cool about it,
>   and correctly recognizes the separate main.

### Neovim

With NvChad[^nvchad] and `gopls`,
add the following to your `$Env:LOCALAPPDATA\nvim\init.lua`,

```lua
vim.lsp.config.gopls = {
  settings = {
    gopls = {
      buildFlags = {"-tags=cli gui"}
    }
  }
}
```

Another way is to add this line to your `$Env:LOCALAPPDATA\nvim\init.lua`
to use the included `.nvim.lua`,

```lua
vim.o.exrc = true
```

[^nvchad]: Probably workable without NvChad, but `gopls` is a must.

### GoLand[^goland]

1.  Open **Settings > Go > Build Tags**,
    then add `cli` and `gui` to **Custom tags**.

2.  Open **Run > Edit Configurations**, then **Add New Configuration** (+).

3.  Check **Use all custom build tags** under Go tool arguments.

[^goland]: See https://www.jetbrains.com/help/go/go-build.html#common.

### VS Code[^vscode]

Use the included VS Code settings,
or manually add the following to your `.vscode/settings.json`,

```json
{
  "gopls.env": {
    "GOFLAGS": "-tags=cli,gui"
  }
}
```

[^vscode]: See https://stackoverflow.com/questions/71790348/gopls-returns-the-error-gopls-no-packages-returned-packages-load-error-for-g/75449491#75449491.

##  TODO

- [ ]   Feature: send specs data to Google Sheets.

