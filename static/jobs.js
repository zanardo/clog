function updateScreen(jobs) {
	document.getElementById('jobs').innerHTML = jobs;
	document.getElementById('jobs').style.opacity = 1.0;
}

function loadJobs() {
	document.getElementById('jobs').style.opacity = 0.5;
	var req = new XMLHttpRequest();
	req.open("GET", "/jobs", true);
	req.onreadystatechange = function() {
		if(req.readyState == 4 && req.status == 200) {
			updateScreen(req.responseText);
		}
	}
	req.send(null);
}

function load() {
	loadJobs();
	window.setInterval(loadJobs, 60000);
}