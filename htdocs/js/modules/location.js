// returns location as text (can be used for the innerHTML of an object).
function watchLocation(update_func) {
	// assume geolocation is disabled by default
	//element.innerHTML = "Waiting to receive position information...";

	if ("geolocation" in navigator) {
		if (window.pacmacro_geo !== undefined)
			navigator.geolocation.clearWatch(window.pacmacro_geo);

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
		//element.innerHTML = "Geolocation is disabled.";
	}
}

export { watchLocation };
