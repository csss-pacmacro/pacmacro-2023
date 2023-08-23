// returns location as text (can be used for the innerHTML of an object).
function watchLocation(element) {
	// assume geolocation is disabled by default
	element.innerHTML = "Waiting to receive position information...";

	if ("geolocation" in navigator) {
		navigator.geolocation.watchPosition(
		// on each update...
		(p) => {
			element.innerHTML = `
				Latitude: ${p.coords.latitude};<br>
				Longitude: ${p.coords.longitude};<br>
				Altitude: ${p.coords.altitude};<br>
				Speed: ${p.coords.speed}.
			`;
		},
		// in the event of an error...
		(e) => {
			element.innerHTML = `Error code ${e}.`
		},
		// watch options
		{
			maximumAge: 0, // don't stop watching
			timeout: 5000,
			enableHighAccuracy: true
		}
		); // watchPosition()
	} else {
		element.innerHTML = "Geolocation is disabled.";
	}
}

export { watchLocation };
