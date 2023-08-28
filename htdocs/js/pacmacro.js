// pacmacro.js

// NOTE: change this in production
var URL_ROOT = "localhost:8080";
var EXPAND_X = 32;
var EXPAND_Y = 32;

// reset pacmacro
function pacmacro_reset() {
	// undefine globals
	window.pacmacro_set_ws = undefined;
	window.pacmacro_ws = undefined;
	window.pacmacro_geo = undefined;
	window.pacmacro_ctx = undefined;
}

// get player ID from cookies
function getID() {
	let ID = "";

	let cookies = document.cookie;
	cookies = cookies.split(';').map(v => v.split('='));

	for (const c of cookies) {
		if (c[0].trim() == "id")
			ID = c[1];
	}

	return ID;
}

// connect to websocket as player
function connectWS(ID, func_recv) {
	// calling function should run connectWS in a try-catch block
	window.pacmacro_ws = new WebSocket(`ws://${URL_ROOT}/api/ws/${ID}`);

	if (window.pacmacro_ws === undefined)
		return false;

	// received API message
	window.pacmacro_ws.addEventListener("message", func_recv);

	return true;
}

// watch location of user; run update_func on each update
function watchLocation(update_func) {
	if ("geolocation" in navigator) {
		// check if watchLocation was previously run;
		// if yes, stop watching location.
		if (window.pacmacro_geo !== undefined)
			navigator.geolocation.clearWatch(window.pacmacro_geo);

		// start watching location
		window.pacmacro_geo = navigator.geolocation.watchPosition(
			// on each update...
			update_func,
			// in the event of an error...
			(e) => { console.log(`watchLocation error: ${e}.`) },
			// watch options
			{
				maximumAge: 0, // don't stop watching
				timeout: 5000,
				enableHighAccuracy: true
			}
		);
	} else {
		return false;
	}

	return true;
}

function stopWatchLocation() {
	if (navigator.geolocation !== undefined &&
		window.pacmacro_geo !== undefined)
		navigator.geolocation.clearWatch(window.pacmacro_geo);

	window.pacmacro_geo = undefined;
}

function convertCoords(map, lat, lon) {
	let plot = {
		x: 0,
		y: 0
	};

	let dlat = map.max.latitude - map.min.latitude;
	let dlon = map.max.longitude - map.min.longitude;

	plot.x = ((lon - map.min.longitude) / dlon) * map.width;
	plot.y = ((lat - map.min.latitude) / dlat) * map.height;

	console.log(`[${lat}, ${lon}] => [${plot.x}, ${plot.y}`);

	return plot;
}

export {
	URL_ROOT,
	EXPAND_X, EXPAND_Y,
	pacmacro_reset,
	getID,
	connectWS,
	watchLocation,
	stopWatchLocation,
	convertCoords
};
