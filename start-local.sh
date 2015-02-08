jsx --release --minify --extension jsx js/ js/ && 
yuicompressor js/uutispuro-19.js > js/uutispuro-19.min.js &&
yuicompressor styles/uutispuro-20.css > styles/uutispuro-20.min.css &&
yuicompressor styles/uutispuro-m-20.css > styles/uutispuro-m-20.min.css &&
yuicompressor styles/uutispuro-small-20.css > styles/uutispuro-small-20.min.css &&
go build && 
./uutispuroReader
