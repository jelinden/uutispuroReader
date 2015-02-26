var newsIntervalId;

function getNews() {
	var xmlhttp = new XMLHttpRequest();
	var url = window.location.pathname + "/ws";

	xmlhttp.onreadystatechange = function() {
	    if (xmlhttp.readyState == 4 && xmlhttp.status == 200) {
	        var obj = JSON.parse(xmlhttp.responseText);
	        addItemsToDom(obj.d.Items, obj.d.Lang);
	    }
	}
	xmlhttp.open("GET", url, true);
	xmlhttp.send();
}

function news() {
    newsIntervalId = setInterval(function() {
        getNews();
    }, 20000);
}

document.addEventListener('DOMContentLoaded', function(){
	getNews();
	news();
});

(function(i,s,o,g,r,a,m){i['GoogleAnalyticsObject']=r;i[r]=i[r]||function(){
(i[r].q=i[r].q||[]).push(arguments)},i[r].l=1*new Date();a=s.createElement(o),
m=s.getElementsByTagName(o)[0];a.async=1;a.src=g;m.parentNode.insertBefore(a,m)
})(window,document,'script','//www.google-analytics.com/analytics.js','ga');
ga('create', 'UA-2171875-4', 'auto');
ga('send', 'pageview');
