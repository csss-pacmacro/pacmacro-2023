import { watchLocation } from "./modules/location.js";
import { URL_ROOT, connectWS } from "./modules/socket.js";

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

window.onload = async () => {
	/* INIT */

	// undefine websockets
	window.pacmacro_ws = undefined;
	window.pacmacro_set_ws = undefined;

	/* LOCATION */

	let location_button = document.getElementById("location-button");
	let lwrite_button = document.getElementById("location-write-button");
	let lclose_button = document.getElementById("location-close-button");
	let location_p = document.getElementById("location");

	let set_ws_open = () => {
		console.log("set_ws_open");
		location_p.innerHTML = "Connected.";

		window.pacmacro_set_ws.addEventListener("message", (e) => {
			location_p.innerHTML = e.data;
		});

		// attempt to log-in
		window.pacmacro_set_ws.send(`{
			"latitude": 0,
			"longitude": 0,
			"command": "password",
			"data": "1234"
		}`);

		// watch location and pass it along to the server
		watchLocation((p) => {
			window.pacmacro_set_ws.send(`{
				"latitude": ${p.coords.latitude},
				"longitude": ${p.coords.longitude},
				"command": "location",
				"data": ""
			}`);
		});
	}

	let set_ws_close = () => {
		console.log("set_ws_close");
		location_p.innerHTML = "Closed.";

		if (window.pacmacro_set_ws === undefined)
			return;

		// try to close connection
		window.pacmacro_set_ws.close();
		window.pacmacro_set_ws = undefined;

		// stop watching location
		if (navigator.geolocation !== undefined &&
			window.pacmacro_geo !== undefined)
			navigator.geolocation.clearWatch(window.pacmacro_geo);
	}

	// wait for user to prompt watching
	location_button.onclick = () => {
		if (window.pacmacro_set_ws !== undefined)
			return; // websocket is already opened

		let ID = getID();

		if (ID.length == 0)
			location_p.innerHTML = "Not registered.";

		location_p.innerHTML = "Connecting...";

		window.pacmacro_set_ws = new WebSocket(`ws://${URL_ROOT}/api/game/set/${ID}`);
		window.pacmacro_set_ws.onopen = set_ws_open;
		window.pacmacro_set_ws.onclose = set_ws_close;
	};

	lwrite_button.onclick = () => {
		if (window.pacmacro_set_ws === undefined)
			return;

		window.pacmacro_set_ws.send(`{
			"latitude": 0,
			"longitude": 0,
			"command": "write",
			"data": ""
		}`);
	}

	lclose_button.onclick = set_ws_close;

	/* REGISTRATION */

	let register_button = document.getElementById("register-button");
	let register_p = document.getElementById("register");

	register_button.onclick = async () => {
		const form = new FormData();

		form.append("type", "2");    // TypeAdmin
		form.append("pass", "1234"); // adminPassword

		let resp;

		try {
			resp = await fetch("/api/player/register", {
				method: "POST",
				body: form
			});
		} catch {
			register_p.innerHTML = "There was an error reaching the server.";
		}

		if (resp.ok) {
			let ID = await resp.text();

			// store ID in cookie
			document.cookie = `id=${ID}`;

			register_p.innerHTML = `Registered Admin with ID ${ID}.`;
		} else
			register_p.innerHTML = "Status code not OK.";
	};

	/* WEBSOCKETS */

	let socket_button = document.getElementById("socket-button");
	let sclose_button = document.getElementById("socket-close-button");
	let socket_p = document.getElementById("socket");

	socket_button.onclick = () => {
		let ID = getID();

		if (ID.length == 0) {
			socket_p.innerHTML = "Player is not registered.";
			return;
		}

		try {
			connectWS(ID, (e) => {
				socket_p.innerHTML = e.data;
			});
		} catch {
			socket_p.innerHTML = "Failed to connect to WS.";
			return;
		}

		sclose_button.style.visibility = "visible";
	};

	sclose_button.style.visibility = "hidden";

	sclose_button.onclick = () => {
		if (window.pacmacro_ws) {
			window.pacmacro_ws.close();
			window.pacmacro_ws = undefined;
		}
		sclose_button.style.visibility = "hidden";
	}
};
