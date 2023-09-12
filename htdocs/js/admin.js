// admin.js
// for admin.html

import {
	NTYPE, NREPS, type, reps
} from "./pacmacro.js";

window.onload = () => {
	let stat = document.getElementById("status");
	let load = document.getElementById("load");
	let list = document.getElementById("list");

	load.onclick = async () => {
		list.innerHTML = ""; // clear list

		let players;
		try {
			players = await fetch("/api/player/list.json");
			if (!players.ok) {
				throw "Response not OK.";
			}

			players = await players.json();
		} catch {
			list.innerHTML = `
				<p>There was a problem contacting the API.</p>
			`;
			return;
		}
		// players is a list of all online players

		for (let i = 0; i < players.length; i++) {
			const p = players[i];

			let types = "";
			// NOTE: -1 is Delete
			for (let i = -1; i < NTYPE; i++) {
				types += `<option value=${i}
					${p.type == i ? "selected" : ""}>
					${type(i)}
				</option>`;
			}

			let reps_ = "";
			for (let i = 0; i < NREPS; i++) {
				reps_ += `<option value=${i}
					${p.reps == i ? "selected" : ""}>
					${reps(i)}
				</option>`;
			}

			list.innerHTML += `<div class="player">
				<h1>${p.name} (${p.id})</h1>
				<select id="type${i}">${types}</select>
				<select id="reps${i}">${reps_}</select>
				<button id="submit${i}">Submit</button>
			</div>`;
			document.getElementById(`submit${i}`).onclick = eval(`
				async () => {
					let id    = document.getElementById("id");
					let pass  = document.getElementById("pass");
					let stat  = document.getElementById("status");
					let _type = document.getElementById("type${i}");
					let _reps = document.getElementById("reps${i}");

					let form = new FormData();

					form.append("id",   id.value);
					form.append("pass", pass.value);
					form.append("type", _type.value);
					form.append("reps", _reps.value);

					try {
						let resp = await fetch("/api/admin/update/${p.id}", {
							method: "POST",
							body: form
						});
						if (!resp.ok) {
							stat.innerHTML = "Response not OK.";
						} else {
							stat.innerHTML = "Success.";
						}
					} catch {
						stat.innerHTML = "Error: couldn't contact API.";
					}
				}
			`);
		}
	};
}
