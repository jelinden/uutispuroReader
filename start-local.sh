## brew install yuicompressor
## npm install -g react-tools
jsx --release --minify --extension jsx assets/js assets/js && 
yuicompressor assets/js/uutispuro.js --type js -o assets/js/uutispuro.min.js && 
yuicompressor assets/js/react.js --type js -o assets/js/react.min.js && 
yuicompressor assets/js/react-0.12.2.js --type js -o assets/js/react-0.12.2.min.js && 
cat assets/js/moment.min.js > assets/js/uutispuro-20.min.js && 
cat assets/js/react-0.12.2.min.js >> assets/js/uutispuro-20.min.js && 
cat assets/js/react.min.js >> assets/js/uutispuro-20.min.js && 
cat assets/js/uutispuro.min.js >> assets/js/uutispuro-20.min.js &&
yuicompressor assets/styles/uutispuro-20.css -o assets/styles/uutispuro-20.min.css &&
yuicompressor assets/styles/uutispuro-m-20.css -o assets/styles/uutispuro-m-20.min.css &&
yuicompressor assets/styles/uutispuro-small-20.css -o assets/styles/uutispuro-small-20.min.css &&
go build && 
./uutispuroReader
