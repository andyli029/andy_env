set nocompatible   " Disable vi-compatibility
set noswapfile
set ts=4
set expandtab
set incsearch
set hlsearch
set showmatch
set backspace=2
set spr " for vs window in right
set cursorline
set encoding=utf-8
set showcmd
set cmdheight=1 " 设置命令行的高度
set autochdir


if has("gui_macvim")
    winpos 20 20            " 指定窗口出现的位置，坐标原点在屏幕左上角
    set lines=60 columns=160 "
    指定窗口大小，lines为高度，columns为宽度
    set guioptions-=m       " 隐藏菜单栏
    set guioptions-=T       " 隐藏工具栏
    set guioptions-=L       " 隐藏左侧滚动条
    set guioptions-=r       " 藏右侧滚动条
    set guioptions+=b       " 显示底部滚动条
    set nowrap              " 设置不自动换行
    set guicursor=i-ci:block-Cursor/lCursor
    set shell=/bin/zsh
endif

set path+=** " search subdir
let mapleader=","
set wildignore+=*.o,*~,*.pyc,*.class,*.swp

call plug#begin('~/.vim/plugged')
" color
" Plug 'altercation/vim-colors-solarized'
Plug 'vim-airline/vim-airline'
Plug 'vim-airline/vim-airline-themes'
Plug 'vim-scripts/Mark'

" file type indent
Plug 'vim-syntastic/syntastic' " sync check
Plug 'editorconfig/editorconfig-vim'
Plug 'Yggdroot/indentLine' " indent display |
Plug 'godlygeek/tabular' " Alignment

Plug 'SirVer/ultisnips' " snip engine
Plug 'honza/vim-snippets' 

Plug 'ervandew/supertab'
Plug 'majutsushi/tagbar' "TagbarToggle

Plug 'scrooloose/nerdcommenter' "Comment

Plug 'airblade/vim-gitgutter' " display git change

Plug 'ctrlpvim/ctrlp.vim'
" Go
Plug 'fatih/vim-go',{ 'do': ':GoInstallBinaries' }


" C/C++
Plug 'Valloric/YouCompleteMe',  { 'do': './install.py --clang-completer' }
Plug 'rdnetto/YCM-Generator', {'branch': 'stable'}
Plug 'vim-scripts/a.vim'
 
" for file type
Plug 'mileszs/ack.vim'
Plug 'tpope/vim-fugitive'
Plug 'elzr/vim-json'
Plug 'pearofducks/ansible-vim'
Plug 'plasticboy/vim-markdown'
call plug#end()

" solarized plugin settings
" set t_Co=256
" set background=dark
" let g:solarized_termcolors=256
" colorscheme solarized

" airline and airline_theme settings
set laststatus=2
let g:airline#extensions#tabline#enabled = 1
let g:airline#extensions#tabline#left_sep = ' '
let g:airline#extensions#tabline#left_alt_sep = '|'
let g:airline#extensions#tabline#formatter = 'default'
let g:airline_theme='term'


" tabular 
if exists(":Tabularize")
    inoremap <silent> <Bar>   <Bar><Esc>:call <SID>align()<CR>a
    function! s:align()
        let p = '^\s*|\s.*\s|\s*$'
        if exists(':Tabularize') && getline('.') =~# '^\s*|' && (getline(line('.')-1) =~# p || getline(line('.')+1) =~# p)
            let column = strlen(substitute(getline('.')[0:col('.')],'[^|]','','g'))
            let position = strlen(matchstr(getline('.')[0:col('.')],'.*|\s*\zs.*'))
            Tabularize/|/l1
            normal! 0
            call search(repeat('[^|]*|',column).'\s\{-\}'.repeat('.',position),'ce',line('.'))
        endif
    endfunction
endif


" vim-go
let g:go_fmt_fail_silently = 1 "sovle vim-go and syntastic_checker conflict
let g:go_list_type = "quickfix"
let g:go_fmt_command = "goimports"
let g:go_def_mode = 'godef'
au FileType go nmap ge <Plug>(go-rename)
au FileType go nmap gv <Plug>(go-def-vertical)
au FileType go nmap gc <Plug>(go-callers)
au FileType go nmap gs <Plug>(go-implements)
au FileType go nmap gl <Plug>(go-lint)
autocmd Filetype go nnoremap <F7> :GoTest<CR>
autocmd FileType qf nmap <buffer> <cr> <cr>:ccl<cr>
"au BufReadPost,BufNewFile *.yml,*.yaml set ft=ansible

" ctrlp
let g:ctrlp_cmd = 'CtrlP'

" plasticboy/vim-markdown
let g:vim_markdown_folding_disabled = 1
let g:vim_markdown_toc_autofit = 1
let g:vim_markdown_json_frontmatter = 1
let g:vim_markdown_new_list_item_indent = 2
let g:vim_markdown_conceal = 0


" editorconfig/editorconfig-vim
let g:EditorConfig_exec_path = "~/.editorconfig"
function! FiletypeHook(config)
    if has_key(a:config, 'vim_filetype')
        let &filetype = a:config['vim_filetype']
    endif
    return 0   " Return 0 to show no error happened
endfunction
call editorconfig#AddNewHook(function('FiletypeHook'))

" vim-syntastic/syntastic
set statusline+=%#warningmsg#
set statusline+=%{SyntasticStatuslineFlag()}
set statusline+=%*
autocmd FileType go let g:syntastic_go_checkers=['go', 'golint']
autocmd FileType yaml let g:syntastic_yaml_checkers=['yamllint']
autocmd FileType lua let g:syntastic_lua_checkers=['luac']
autocmd FileType vim let g:syntastic_vim_checkers=['Vint']
autocmd FileType ansible let g:syntastic_ansible_checkers=['ansible_lint']
autocmd FileType markdown let g:syntastic_markdown_checkers=['proselint']

let g:syntastic_always_populate_loc_list = 1
let g:syntastic_error_symbol='✗'
let g:syntastic_warning_symbol='⚠'

" ansible-vim
let g:ansible_name_highlight = 'd'
let g:ansible_unindent_after_newline = 1

" ervandew/supertab
let g:SuperTabRetainCompletionType=2
let g:SuperTabClosePreviewOnPopupClose=1

" ################### 自动补全 ###################
" set completeopt+=preview
let g:ycm_min_num_of_chars_for_completion = 3
let g:ycm_add_preview_to_completeopt=1
let g:ycm_autoclose_preview_window_after_insertion = 1
let g:ycm_autoclose_preview_window_after_completion=1
let g:ycm_complete_in_comments = 1
let g:ycm_key_list_select_completion = ['<c-j>', '<Down>']
let g:ycm_key_list_previous_completion = ['<c-k>', '<Up>']
let g:ycm_confirm_extra_conf = 0
let g:ycm_disable_for_files_larger_than_kb = 5000
nnoremap gd :YcmCompleter GoToDefinitionElseDeclaration<CR>
autocmd Filetype c,cpp,objc,objcpp,javascript,typescript,java nnoremap <leader><space> :YcmCompleter GetType <C-R>=expand("<cword>")<CR><CR>
autocmd Filetype c,cpp,objc,objcpp nnoremap gi :YcmCompleter GoToInclude <C-R>=expand("<cword>")<CR><CR>
autocmd Filetype c,cpp,objc,objcpp,javascript,typescript,java nnoremap gt :YcmCompleter GetType <C-R>=expand("<cword>")<CR><CR>
autocmd Filetype c,cpp,objc,objcpp,cs,python,typescript,javascript,rust,java nnoremap gf :YcmCompleter GetDoc <C-R>=expand("<cword>")<CR><CR>
au Filetype c,cpp,objc,objcpp,cs,java nnoremap <leader>f :YcmCompleter FixIt<CR>

" majutsushi/tagbar
let tagbar_left=1
let g:tagbar_sort = 0

" nerdcommenter settings
" Enable trimming of trailing whitespace when uncommenting
let g:NERDTrimTrailingWhitespace = 1
let g:NERDSpaceDelims = 1
let g:NERDDefaultAlign = 'left'

nmap <Leader>s :Ack!<Space><cword><cr>

"
" 比较喜欢用tab来选择补全...
function! MyTabFunction ()
    let line = getline('.')
    let substr = strpart(line, -1, col('.')+1)
    let substr = matchstr(substr, "[^ \t]*$")
    if strlen(substr) == 0
        return "\<tab>"
    endif
    return pumvisible() ? "\<c-n>" : "\<c-x>\<c-o>"
endfunction
inoremap <tab> <c-r>=MyTabFunction()<cr>

" Yggdroot/indentLine
let g:indentLine_enabled = 1
let g:indentLine_color_term = 239

" none X terminal
let g:indentLine_color_tty_light = 7 " (default: 4)
let g:indentLine_color_dark = 1 " (default: 2)
let g:indentLine_setColors = 1
let g:indentLine_concealcursor = 'inc'
let g:indentLine_conceallevel = 2
let g:indentLine_setConceal = 1
let g:indentLine_enabled = 1
nnoremap <leader>il :IndentLinesToggle<CR>

" UltiSnips-triggers
let g:UltiSnipsExpandTrigger="<tab>"
let g:UltiSnipsJumpForwardTrigger="<tab>"

" F1 废弃这个键,防止调出系统帮助
" I can type :help on my own, thanks.  Protect your fat fingers from the evils of <F1>
map <F1> <Esc>"

" F2 行号开关，用于鼠标复制代码用
" 为方便复制，用<F2>开启/关闭行号显示:
function! HideNumber()
	if(&relativenumber == &number)
		set relativenumber! number!
	elseif(&number)
		set number!
	else
		set relativenumber!
	endif
	set number?
endfunc
nnoremap <F2> :call HideNumber()<CR>
nnoremap <F3> :TagbarToggle<CR>
set pastetoggle=<F4>            "    when in insert mode, press <F4> to go to

" disbale paste mode when leaving insert mode
au InsertLeave * set nopaste

map <Left> <Nop>
map <Right> <Nop>
map <Up> <Nop>
map <Down> <Nop>
" 分屏窗口移动, Smart way to move between windows
map <C-j> <C-W>j
map <C-k> <C-W>k
map <C-h> <C-W>h
map <C-l> <C-W>l

" 命令行模式增强，ctrl - a到行首， -e 到行尾
cnoremap <C-j> <t_kd>
cnoremap <C-k> <t_ku>
cnoremap <C-a> <Home>
cnoremap <C-e> <End>
noremap H ^
noremap L $
nnoremap gu gU
nnoremap gl gu
" Quickly close the current window
nnoremap <leader>q :q<CR>

" Quickly save the current file
nnoremap <leader>w :w<CR>
nnoremap <space> za

autocmd InsertEnter * let mapleader = ""
