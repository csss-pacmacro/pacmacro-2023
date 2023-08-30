// pacmacro.js
// general programming for all pages

var WS       = "ws"; // change to "wss" in production
var URL_ROOT = "localhost:8080"; // must be root domain of server hosting API
var EXPAND_X = 32;
var EXPAND_Y = 32;

/* GLOBALS
 * window.pacmacro_set_ws  : Admin; setting map minimum and maximum latitude and longitude values
 * window.pacmacro_ws      : websocket connection to API for lobby or PacMacro game
 * window.pacmacro_geo     : geolocation value associated with current watch function
 * window.pacmacro_ctx     : canvas API 2d context for drawing PacMacro game
 * window.pacmacro_map     : PacMacro map information (JSON; /api/game/map.json)
 */

// reset pacmacro
function pacmacro_reset() {
	// undefine globals
	window.pacmacro_set_ws = undefined;
	window.pacmacro_ws     = undefined;
	window.pacmacro_geo    = undefined;
	window.pacmacro_ctx    = undefined;
	window.pacmacro_map    = undefined;
}

// save ID in cookies
function saveID(ID) {
	document.cookie = `id=${ID}`;
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
function connectWS(ID, onopen, onclose, onerror, onmessage) {
	// calling function should run connectWS in a try-catch block
	window.pacmacro_ws = new WebSocket(`${WS}://${URL_ROOT}/api/ws/${ID}`);
	window.pacmacro_ws.onopen = onopen;
	window.pacmacro_ws.onclose = onclose;
	window.pacmacro_ws.onerror = onerror;
	window.pacmacro_ws.onmessage = onmessage;
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
	WS,
	URL_ROOT,
	EXPAND_X, EXPAND_Y,
	pacmacro_reset,
	saveID,
	getID,
	connectWS,
	watchLocation,
	stopWatchLocation,
	convertCoords
};
