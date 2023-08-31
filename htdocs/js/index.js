// index.js
// programming for the lobby and PacMacro game

import {
	pacmacro_reset,
	connectWS,
	watchLocation
} from "./pacmacro.js";

window.onload = () => {
	// reset globals
	pacmacro_reset();

	let pacmacro_status = document.getElementById("pacmacro-status");
	let pacmacro_canvas = document.getElementById("pacmacro-canvas");
	let pass = ""; // received upon window.prompt(...)

	let pacmacro_open = async () => {
		// send password to authenticate player
		window.pacmacro_ws.send(pass);

		try {
			let pacmacro_map = await fetch("/api/game/map.json");
			window.pacmacro_map = await pacmacro_map.json();
		} catch {
			pacmacro_status.innerHTML = "Error: Couldn't receive map information.";
			window.pacmacro_ws.close(); // close websocket connection
			return; // do not proceed
		}

		console.log(window.pacmacro_map);

		window.pacmacro_ctx = pacmacro_canvas.getContext("2d");
		// set canvas size as per window.pacmacro_map

		watchLocation((p) => {
			let coordinate = {
				"latitude": p.coords.latitude,
				"longitude": p.coords.longitude
			};

			window.pacmacro_ws.send(JSON.stringify(coordinate));
		});

		pacmacro_status.innerHTML = "Good to go.";
	};

	let pacmacro_redirect = () => {
		window.location.href = "/register";
	};

	let pacmacro_recv = (e) => {
		// write server message to status element
		pacmacro_status.innerHTML = e.data;
	};

	const params = new URLSearchParams(window.location.search);

	if (!params.has("id"))
		pacmacro_redirect();
	else {
		pass = window.prompt(`Please enter your password`, "");

		connectWS(params.get("id"),
			pacmacro_open, // on open
			pacmacro_redirect, // on close
			pacmacro_redirect, // on error
			pacmacro_recv); // on message
	}
}
