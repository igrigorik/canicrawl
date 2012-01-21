
desc "Generate ronn file"
task :docs do
	exec('cat README.md | ronn -5 -f --style 80c --pipe > static/canicrawl.1.html')
end
