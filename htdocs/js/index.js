// index.js
// programming for the lobby and PacMacro game

import {
	ribbons,
	EXPAND_X, EXPAND_Y,
	pacmacro_init,
	connectWS,
	watchLocation,
	convertCoords,
	reps,
} from "./pacmacro.js";

window.onload = async () => {
	ribbons();
	// reset globals
	pacmacro_init();

	let pacmacro_status = document.getElementById("pacmacro-status");
	let pacmacro_canvas = document.getElementById("pacmacro-canvas");
	let pass = ""; // received upon window.prompt(...)
	let self_summary = document.getElementById("self-summary");

	try {
		let pacmacro_map = await fetch("/api/game/map.json");
		window.pacmacro_map = await pacmacro_map.json();
	} catch {
		pacmacro_status.innerHTML = "Error: Couldn't receive map information.";
		window.pacmacro_ws.close(); // close websocket connection
		return; // do not proceed
	}

	window.pacmacro_ctx = pacmacro_canvas.getContext("2d");
	// set canvas size as per window.pacmacro_map
	window.pacmacro_ctx.canvas.width = window.pacmacro_map.width * EXPAND_X;
	window.pacmacro_ctx.canvas.height = window.pacmacro_map.height * EXPAND_Y;
	window.pacmacro_ctx.font = "16pt sans-serif";

	let pacmacro_draw = () => {
		if (window.pacmacro_ctx === undefined)
			return;

		// fill canvas with a black rectangle
		window.pacmacro_ctx.fillStyle = "#000000";
		window.pacmacro_ctx.fillRect(0, 0,
			window.pacmacro_ctx.canvas.width,
			window.pacmacro_ctx.canvas.height);

		// render each player
		Object.keys(window.pacmacro_players).forEach((ID,i) => {
			const p = window.pacmacro_players[ID];
			let img, text;

			switch (p.player.reps) {
			case 2: // pacman
				img = window.pacmacro_img_pacman;
				break;
			case 3: // ghost
				img = window.pacmacro_img_ghost;
				break;
			default: // nothing; watcher; error
				return; // do not render image
			}

			window.pacmacro_ctx.drawImage(img,
				p.plot.x * EXPAND_X - 48,
				p.plot.y * EXPAND_Y - 96,
				96, 96);
			window.pacmacro_ctx.textAlign = "center";
			window.pacmacro_ctx.fillStyle = "#ffffff";

			if (ID == getID()) {
				text = `${p.player.name} (You)`;
			} else {
				text = `${p.player.name} (${ID})`;
			}

			window.pacmacro_ctx.fillText(text,
				p.plot.x * EXPAND_X,
				p.plot.y * EXPAND_Y - 112);
		});
	};

	let pacmacro_open = async () => {
		// send password to authenticate player
		window.pacmacro_ws.send(pass);

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
		window.location.href = "/login";
	};

	let pacmacro_recv = (e) => {
		let msg = JSON.parse(e.data);
		console.log(msg.command);

		if (msg.command === undefined)
			return; // invalid message

		if (msg.command == "update-self") {
			window.pacmacro_self = JSON.parse(msg.data);
			self_summary.innerHTML =
				`${window.pacmacro_self.name} is ${reps(window.pacmacro_self.reps)}.`;
		} else
		if (msg.command == "inform") {
			let plot = convertCoords(window.pacmacro_map,
				msg.coordinate.latitude, msg.coordinate.longitude);
			const player = JSON.parse(msg.data);
			console.log(player);

			if (player.id === undefined)
				return; // invalid

			// set player
			window.pacmacro_players[player.id] = {
				"plot": plot,
				"player": player
			};
		} else
		if (msg.command == "move") {
			let plot = convertCoords(window.pacmacro_map,
				msg.coordinate.latitude, msg.coordinate.longitude);
			const p = window.pacmacro_players[msg.data];

			if (p === undefined)
				return; // invalid

			// update plot
			p.plot = plot;
		}

		// write server message to status element
		pacmacro_status.innerHTML = e.data;
		pacmacro_draw(); // re-draw the map; maybe there was a player update
	};

	const params = new URLSearchParams(window.location.search);

	if (!params.has("id"))
		pacmacro_redirect();
	else {
		let ID = params.get("id");
		pass = window.prompt(`Please enter ${ID}'s password`, "");

		connectWS(ID,
			pacmacro_open, // on open
			pacmacro_redirect, // on close
			pacmacro_redirect, // on error
			pacmacro_recv); // on message
	}
}
