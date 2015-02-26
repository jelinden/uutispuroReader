var NewsItems = React.createClass({displayName: "NewsItems",
	handleClick: function(id, eventType) {
		if (id != undefined) {
			var xmlhttp = new XMLHttpRequest();
			if (eventType == 'click') {
				xmlhttp.open("GET", "/c/" + id, true);
				xmlhttp.send();
			} else if (eventType == 'like' && document.getElementById(id).getElementsByClassName("like").length > 0) {
				xmlhttp.open("GET", "/l/" + id, true);	
				xmlhttp.send();
	      		document.getElementById(id).getElementsByClassName("like")[0].innerHTML = (+document.getElementById(id).getElementsByClassName("like")[0].innerHTML + 1);
				document.getElementById(id).getElementsByClassName("like")[0].className = "boldlike";
			} else if (eventType == 'unlike' && document.getElementById(id).getElementsByClassName("unlike").length > 0) {
				xmlhttp.open("GET", "/u/" + id, true);
				xmlhttp.send();
	      		document.getElementById(id).getElementsByClassName("unlike")[0].innerHTML = (+document.getElementById(id).getElementsByClassName("unlike")[0].innerHTML + 1);
				document.getElementById(id).getElementsByClassName("unlike")[0].className = "boldunlike";				
			}
			
		}
	},
	render: function() {
		{var newsImg = ''}
		{var lang = this.props.lang}
	    return (
			React.createElement("div", null, 
	        this.props.rssItems.map(function(result) {
				{if(result.Enclosure.Url != '') {
					newsImg = React.createElement("img", {className: "imgsize", onerror: "this.src=\\\"\\\"", src: result.Enclosure.Url})
				} else {
					newsImg = ''
				}
				var blackBackground = (result.Source==='Turun Sanomat'||result.Source==='Telegraph')?'':'black';
				}
	          	return React.createElement("ul", {key: result.id}, 
					React.createElement("li", {className: "first"}, 
						React.createElement("div", {className: 'img '+blackBackground}, 
							newsImg
						)
					), 
					React.createElement("li", {className: "second"}, 
						React.createElement("div", {className: "source"}, result.Source), 
						React.createElement("div", {className: 'category ' + result.Category.StyleName}, 
							React.createElement("a", {href: '/category/' + result.Category.Name + '/' + window.location.pathname.split('/')[1] + '/'}, categoryName(lang, result.Category.Name))
						), 
						React.createElement("div", {className: "date"}, 
							moment(result.Date).format("DD.MM. HH:mm")
						), 
						React.createElement("div", {className: "social", id: result.id}, 
							React.createElement("span", {className: "like", onClick: this.handleClick.bind(this, result.id, 'like')}, result.Likes), 
							React.createElement("span", {className: "unlike", onClick: this.handleClick.bind(this, result.id, 'unlike')}, result.Unlikes)
						), 
						React.createElement("div", {className: "link"}, 
							React.createElement("a", {id: result.id, target: "_blank", onClickCapture: this.handleClick.bind(this, result.id, 'click'), href: result.Link}, result.Title)
						)
					)
				);
			}, this)
			)
	    );
	}}
);
	
function addItemsToDom(rssItems, lang) {
	React.render(React.createElement(NewsItems, {rssItems: rssItems, lang: lang}), document.getElementById('news-container'));
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
