
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
    }, 5000);
}

function openWebSocket() {
    var $ul = $('#news-container');
    ws = new WebSocket("ws://localhost:9100/websocket");

    ws.onmessage = function(event) {            
        var items = [];
        var obj = $.parseJSON(event.data);
        $.each(obj.d, function(count, rss) {
            items.push("<ul>");
            if(rss.Enclosure.Url != '') {
                items.push("<li class='first'><img width='120' src='" + rss.Enclosure.Url + "'/></li>");
            } else {
                items.push("<li class='first'><span class='img'>&nbsp;</span></li>");
            }
            items.push("<li>" + $.format.date(rss.Date, 'dd.MM. hh:mm') + " " + rss.Source + "(" + rss.Category.Name + ")</li>");
            items.push("<li><a target='_blank' href='" + rss.Link + "'>" + rss.Title + "</a></li>");
            items.push("</ul>");
        });

        $ul.replaceWith(
            $("<div/>", {
                "class": "news-list",
                html: items.join("")
            })
        );
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
}

function news() {
    newsIntervalId = setInterval(function() {
        if (ws != undefined) {
    	    ws.send("ping");
        } else {
            clearInterval(newsIntervalId);
        }
    },5000);
}

$(function() {
    openWebSocket();
    $('body').click(function(e){ console.log("click " + e.target) })  
});