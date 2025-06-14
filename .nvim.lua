vim.lsp.config.gopls = {
	settings = {
		gopls = {
			buildFlags = {"-tags=cli gui"}
		}
	}
}

if type(vim.cmd.OutlineOpen) == 'function' then
	vim.cmd.OutlineOpen()
end

