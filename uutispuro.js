
var ws;
var newsIntervalId;
var connectionIntervalId;

function retryOpeningWebSocket() {
    $('#status').html('Trying to reconnect');
    connectionIntervalId = setInterval(function() {
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
    var $ul = $('#news-container');
    if (ws == undefined || ws == null) {
        ws = new WebSocket(url());
    }

    ws.onmessage = function(event) {
        var items = [];
        var obj = $.parseJSON(event.data);
        $.each(obj.d, function(count, rss) {

            if($('#' + rss.id).html() == undefined) {
                items.push("<ul id='" + rss.id + "'>");
                if(rss.Enclosure.Url != '') {
                    items.push("<li class='first'><img width='120' src='" + rss.Enclosure.Url + "'/></li>");
                } else {
                    items.push("<li class='first'><span class='img'>&nbsp;</span></li>");
                }
                items.push("<li>" + $.format.date(rss.Date, 'dd.MM. HH:mm') + " " + rss.Source + "(" + rss.Category.Name + ")</li>");
                items.push("<li><a target='_blank' href='" + rss.Link + "'>" + rss.Title + "</a></li>");
                items.push("</ul>");
            }
        });

        if (!$.isEmptyObject(items)) {
            var itemslength = items.length;
            $ul.prepend(items.join(""));
            
            if ($ul.length > 30) {
                $ul.find("ul:nth-last-child(-n+" + itemslength + ")").remove();
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
        $('#status', {"class": "bg-success"});
        $('#status').html('Socket open');
        clearInterval(connectionIntervalId);
        ws.send("ping");
        news();
    };
    ws.onerror = function(event) {
        $('#error').html('error ' + evt.toString());;
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
    $('body').click(function(e){ console.log("click " + e.target) })  
});