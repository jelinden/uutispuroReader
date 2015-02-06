'use strict';

var ws;
var newsIntervalId;
var connectionIntervalId;
var connectionRetryCounter = 0;

function retryOpeningWebSocket() {
    connectionIntervalId = setInterval(function() {
        if (connectionRetryCounter == 2) {
            location.reload(true);
        }
        connectionRetryCounter +=1;
        openWebSocket();
        if (ws != undefined && ws.readyState === ws.OPEN) {
            clearInterval(connectionIntervalId);
        }
    }, 1100);
}

function url() {
    var l = window.location;
    return ((l.protocol === "https:") ? "wss://" : "ws://") +
        l.hostname + (((l.port != 80) && (l.port != 443)) ? ":" + l.port : "") + 
		"/websocket" + l.pathname;
}

function openWebSocket() {
  if ("WebSocket" in window) {
    if (ws == undefined || ws == null) {
        ws = new WebSocket(url());
    }

    ws.onmessage = function(event) {
		var obj = JSON.parse(event.data)
		addItemsToDom(obj.d.Items, obj.d.Lang);
    };
    ws.onclose = function (event) {
        clearInterval(newsIntervalId);
        retryOpeningWebSocket();
    };
    ws.onopen = function (event) {
        clearInterval(connectionIntervalId);
        ws.send("ping");
        news();
    };
    ws.onerror = function(event) {

    };
  } else if (!("WebSocket" in window)) {
    alert("WebSocket NOT supported by your Browser! Please change to a modern browser.");
  }
}

function news() {
    newsIntervalId = setInterval(function() {
        if (ws != undefined && ws.readyState === ws.OPEN) {
    	    ws.send("ping");
        }
    },20000);
}

var NewsItems = React.createClass({
	handleClick: function(id, isLike) {
		if (ws != undefined && ws.readyState === ws.OPEN && id != undefined) {
			if (isLike && document.getElementById(id).getElementsByClassName("like").length > 0) {		
				ws.send("l/" + id);
	      		document.getElementById(id).getElementsByClassName("like")[0].innerHTML = (+document.getElementById(id).getElementsByClassName("like")[0].innerHTML + 1);
				document.getElementById(id).getElementsByClassName("like")[0].className = "boldlike";
			} else if (document.getElementById(id).getElementsByClassName("unlike").length > 0) {
				ws.send("u/" + id);
	      		document.getElementById(id).getElementsByClassName("unlike")[0].innerHTML = (+document.getElementById(id).getElementsByClassName("unlike")[0].innerHTML + 1);
				document.getElementById(id).getElementsByClassName("unlike")[0].className = "boldunlike";				
			}
			
		}
	},
	render: function() {
		{var newsImg = ''}
		{var lang = this.props.lang}
	    return (
			<div>
	        {this.props.rssItems.map(function(result) {
				{if(result.Enclosure.Url != '') {
					newsImg = <img className='imgsize' onerror='this.src=\"\"' src={result.Enclosure.Url}/>
				} else {
					newsImg = ''
				}
				var blackBackground = (result.Source==='Turun Sanomat'||result.Source==='Telegraph')?'':'black';
				}
	          	return <ul key={result.id}>
					<li className='first'>
						<div className={'img '+blackBackground}>
							{newsImg}
						</div>
					</li>
					<li className='second'>
						<div className='source'>{result.Source}</div>
						<div className={'category '+result.Category.StyleName}>
							<a href={'/'+window.location.pathname.split('/')[1]+'/category/'+result.Category.Name}>{categoryName(lang, result.Category.Name)}</a>
						</div>
						<div className='date'>
							{moment(result.Date).format("DD.MM. HH:mm")}
						</div>
						<div className='social' id={result.id}>
							<span className='like' onClick={this.handleClick.bind(this, result.id, true)}>{result.Likes}</span>
							<span className='unlike' onClick={this.handleClick.bind(this, result.id, false)}>{result.Unlikes}</span>
						</div>
						<div className='link'>
							<a id={result.id} target='_blank' href={result.Link}>{result.Title}</a>
						</div>
					</li>
				</ul>;
			}, this)}
			</div>
	    );
	}}
);
	
function addItemsToDom(rssItems, lang) {
	React.render(<NewsItems rssItems={rssItems} lang={lang}/>, document.getElementById('news-container'));
}

function categoryName(lang, cat) {
    if (lang === 2) {
        if (cat === 'Digi') {
            return 'Tech';
        } else if (cat === 'Elokuvat') {
            return 'Movies';
        } else if (cat === 'Koti') {
            return 'Home';
        } else if (cat === 'Kotimaa') {
            return 'Domestic';
        } else if (cat === 'Kulttuuri') {
            return 'Culture';
        } else if (cat === 'Matkustus') {
            return 'Travel';
        } else if (cat === 'Pelit') {
            return 'Games';
        } else if (cat === 'Ruoka') {
            return 'Food';
        } else if (cat === 'Talous') {
            return 'Economy';
        } else if (cat === 'Terveys') {
            return 'Health';
        } else if (cat === 'Tiede') {
            return 'Science';
        } else if (cat === 'Ulkomaat') {
            return 'Foreign';
        } else if (cat === 'Urheilu') {
            return 'Sports';
        } else if (cat === 'Viihde') {
            return 'Entertainment';
        } else if (cat === 'Blogit') {
            return 'Blogs';
        } else if (cat === 'Naiset') {
            return 'Women';
        }
    } else if (lang === '/sv/') {
    } else {
        return cat;
    }
}

document.addEventListener('DOMContentLoaded',function(){
	openWebSocket();	
});

(function(i,s,o,g,r,a,m){i['GoogleAnalyticsObject']=r;i[r]=i[r]||function(){
(i[r].q=i[r].q||[]).push(arguments)},i[r].l=1*new Date();a=s.createElement(o),
m=s.getElementsByTagName(o)[0];a.async=1;a.src=g;m.parentNode.insertBefore(a,m)
})(window,document,'script','//www.google-analytics.com/analytics.js','ga');
ga('create', 'UA-2171875-4', 'auto');
ga('send', 'pageview');
