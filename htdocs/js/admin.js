// admin.js

import {
	URL_ROOT,
	pacmacro_reset,
	getID,
	connectWS,
	watchLocation
} from "./pacmacro.js";

window.onload = async () => {
	pacmacro_reset();

	/* ADMIN INFORMATION */

	let admin_id   = document.getElementById("admin-id");
	let admin_pass = document.getElementById("admin-pass");

	document.getElementById("load-id").onclick = () => {
		admin_id.value = getID();
	}

	/* REGISTRATION */

	let register_button  = document.getElementById("register-button");
	let register_p       = document.getElementById("register-status");

	register_button.onclick = async () => {
		const form = new FormData();

		form.append("type", "2"); // TypeAdmin
		form.append("pass", admin_pass.value);

		let resp;

		// attempt to register admin
		try {
			resp = await fetch("/api/player/register", {
				method: "POST",
				body: form
			});
		} catch {
			// on fetch error
			register_p.innerHTML = "Error";
		}

		if (resp.ok) {
			let ID = await resp.text();

			// store ID in cookie
			document.cookie = `id=${ID}`;

			register_p.innerHTML = ID;
		} else
			// on form error
			register_p.innerHTML = `${resp.status}`;
	};

	/* LOCATION */

	let lopen_button  = document.getElementById("location-open-button");
	let lwrite_button = document.getElementById("location-write-button");
	let lclose_button = document.getElementById("location-close-button");
	let location_p    = document.getElementById("location-status");

	// on connection opened
	let set_ws_open = () => {
		location_p.innerHTML = "Connected.";

		// on server message
		window.pacmacro_set_ws.addEventListener("message", (e) => {
			// show server message in status
			location_p.innerHTML = e.data;
		});

		// send log-in information
		window.pacmacro_set_ws.send(
// -------------<
`{
	"coordinate": {
		"latitude": 0,
		"longitude": 0
	},
	"command": "password",
	"data": "${admin_pass.value}"
}`
// -------------<
		);

		// watch location and pass it along to the server
		watchLocation((p) => {
			window.pacmacro_set_ws.send(
// ---------------------<
`{
	"coordinate": {
		"latitude": ${p.coords.latitude},
		"longitude": ${p.coords.longitude}
	},
	"command": "location",
	"data": ""
}`
// ---------------------<
			);
		});
	} // set_ws_open

	// on connection closed
	let set_ws_close = () => {
		location_p.innerHTML = "Closed.";

		// stop watching location
		if (navigator.geolocation !== undefined &&
			window.pacmacro_geo !== undefined)
			navigator.geolocation.clearWatch(window.pacmacro_geo);

		if (window.pacmacro_set_ws !== undefined) {
			// close connection ourselves (on lclose_button.onclick())
			window.pacmacro_set_ws.close();
			window.pacmacro_set_ws = undefined;
		}
	} // set_ws_close

	// on prompt to start map setting
	lopen_button.onclick = () => {
		if (window.pacmacro_set_ws !== undefined)
			return; // websocket is already opened

		location_p.innerHTML = "Connecting...";

		// try to open connection
		window.pacmacro_set_ws = new WebSocket(`ws://${URL_ROOT}/api/admin/set/${admin_id.value}`);
		window.pacmacro_set_ws.onopen = set_ws_open;
		window.pacmacro_set_ws.onclose = set_ws_close;
	}; // location_button.onclick

	// on prompt to write collected location data to server
	lwrite_button.onclick = () => {
		if (window.pacmacro_set_ws === undefined)
			return; // there isn't an open connection

		// send command to write data
		window.pacmacro_set_ws.send(
// -------------<
`{
	"coordinate": {
		"latitude": 0,
		"longitude": 0
	},
	"command": "write",
	"data": ""
}`
// -------------<
		);
	}; // lwrite_button.onclick

	lclose_button.onclick = set_ws_close;
};
