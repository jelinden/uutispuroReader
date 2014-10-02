
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
    }, 30000);
}

function url() {
    var l = window.location;
    return ((l.protocol === "https:") ? "wss://" : "ws://") + 
        l.hostname + (((l.port != 80) && (l.port != 443)) ? ":" + l.port : "") + "/websocket";
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
        if ($('#news-container ul').length > 0) {
            mintbg = 'mint';
        }
        var rssItems = obj.d;
        $.each(rssItems, function(count, rss) {
            if($('#' + rss.id).html() == undefined) {
                items.push("<ul id='" + rss.id + "' class='hiddenelement " + mintbg + "'>");
                if(rss.Enclosure.Url != '') {
                    items.push("<li class='first img'><img src='" + rss.Enclosure.Url + "'/></li>");
                } else {
                    items.push("<li class='first img'>&nbsp;</li>");
                }
                var category = rss.Category.Name == 'IT ja media'?'Digi':rss.Category.Name
                items.push("<li class='second'><div class='source'>" + rss.Source + "</div><div class='category " + category + "'>" + category + "</div><div class='date'>" + $.format.date(rss.Date, 'dd.MM. HH:mm') + "</div>");
                items.push("<div class='link'><a target='_blank' href='" + rss.Link + "'>" + rss.Title + "</a></div></li>");
                items.push("</ul>");
            }
        });

        if (!$.isEmptyObject(items)) {
            $('#news-container ul').removeClass("mint");
            $ul.prepend(items.join(""));
            $(".hiddenelement").fadeIn(2500);
            var containerLength = $('#news-container ul').length;
            if (containerLength > 40) {
                $ul.find("ul:nth-last-child(-n+" + (containerLength-39)  + ")").remove();
            }
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
        $('#error').html('error ' + event.toString());
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

$(function() {
    openWebSocket();
    $('body').click(function(e){ /*console.log("click " + e.target)*/ })  
});