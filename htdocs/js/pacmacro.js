// pacmacro.js

// NOTE: change this in production
var URL_ROOT = "localhost:8080";

// reset pacmacro
function pacmacro_reset() {
	// undefine globals
	window.pacmacro_set_ws = undefined;
	window.pacmacro_ws = undefined;
	window.pacmacro_geo = undefined;
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

export {
	URL_ROOT,
	pacmacro_reset,
	getID,
	connectWS,
	watchLocation
};
