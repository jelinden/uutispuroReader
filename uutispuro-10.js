
var ws;
var newsIntervalId;
var connectionIntervalId;
var connectionRetryCounter = 1;
function retryOpeningWebSocket() {
    $('#status').html('Trying to reconnect');
    connectionIntervalId = setInterval(function() {
        if (connectionRetryCounter == 3) {
            location.reload(true);
        }
        connectionRetryCounter +=1;
        openWebSocket();
        if (ws != undefined && ws.readyState === ws.OPEN) {
            $('#status').html('Reconnected');
            clearInterval(connectionIntervalId);
        }
    }, 1500);
}

function url() {
    var l = window.location;
    return ((l.protocol === "https:") ? "wss://" : "ws://") +
        l.hostname + (((l.port != 80) && (l.port != 443)) ? ":" + l.port : "") + "/websocket" + l.pathname;
}

function openWebSocket() {
  if ("WebSocket" in window) {
    if (ws == undefined || ws == null) {
        ws = new WebSocket(url());
    }

    ws.onmessage = function(event) {
        var $ul = $('#news-container');
        var items = [];
        var mintbg = '';
        var obj = $.parseJSON(event.data);
        if ($('#news-container ul').length > 0 && $('#news-container ul').length != 12) {
            mintbg = 'mint';
        }
        var rssItems = obj.d.Items;
        $.each(rssItems, function(count, rss) {
            if($('#' + rss.id).html() == undefined) {
                items.push("<ul id='" + rss.id + "' class='hiddenelement " + mintbg + "'>");
                var blackBackground = (rss.Source==='Turun Sanomat'||rss.Source==='Telegraph')?'':'black';
                if(rss.Enclosure.Url != '') {
                    items.push("<li class='first'><div class='img " + blackBackground + "'><img src='" + rss.Enclosure.Url + "'/></div></li>");
                } else {
                    items.push("<li class='first'><div class='img " + blackBackground + "'>&nbsp;</div></li>");
                }
                var category = rss.Category.Name;

                items.push("<li class='second'><div class='source'>" + rss.Source + "</div><div class='category " + rss.Category.StyleName + "'>" + categoryName(obj.d.Lang, category) + "</div><div class='date'>" + $.format.date(rss.Date, 'dd.MM. HH:mm') + "</div>");
                items.push("<div class='link'><a id='" + rss.id + "' target='_blank' href='" + rss.Link + "'>" + rss.Title + "</a></div></li>");
                items.push("</ul>");
            }
        });
        if (!$.isEmptyObject(items)) {
            $('#news-container ul').removeClass("mint");
            if ($ul.find("ul").length == 12) {
                $ul..append()(items.join(""));
            } else {
                $ul.prepend(items.join(""));
            }
            $(".hiddenelement").fadeIn(2500);
            
            var containerLength = $('#news-container ul').length;
            if (containerLength > 45) {
                $ul.find("ul:nth-last-child(-n+" + (containerLength-45)  + ")").remove();
            }
            $('#news-container ul').removeClass("hiddenelement");
        }
    };
    ws.onclose = function (event) {
        $('#status', {"class": "bg-warning"});
        $('#status').html('Socket closed');
        clearInterval(newsIntervalId);
        retryOpeningWebSocket();
    };
    ws.onopen = function (event) {
        $('#status', {"class": ""});
        $('#status').html('');
        clearInterval(connectionIntervalId);
        ws.send("ping");
        news();
    };
    ws.onerror = function(event) {

    };
  } else {
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

function categoryName(lang, cat) {
    if (lang === 2) {
        if (cat === 'IT ja media') {
            return 'Digital media';
        } else if (cat === 'Digi') {
            return 'Digital media';
        } else if (cat === 'TV ja elokuvat') {
            return 'TV and movies';
        } else if (cat === 'Asuminen') {
            return 'Home and living';
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
        } else if (cat === 'Naiset ja muoti') {
            return 'Women and fashion';
        }
    } else if (lang === '/sv/') {

    } else {
        return cat;
    }
}

$(function() {
    openWebSocket();
    $(document).delegate('.link a', 'click', function(e) {
        if (ws != undefined && ws.readyState === ws.OPEN) {
            ws.send("c/" + e.target.id);
        }
    })
});

(function(i,s,o,g,r,a,m){i['GoogleAnalyticsObject']=r;i[r]=i[r]||function(){
(i[r].q=i[r].q||[]).push(arguments)},i[r].l=1*new Date();a=s.createElement(o),
m=s.getElementsByTagName(o)[0];a.async=1;a.src=g;m.parentNode.insertBefore(a,m)
})(window,document,'script','//www.google-analytics.com/analytics.js','ga');
ga('create', 'UA-2171875-4', 'auto');
ga('send', 'pageview');
