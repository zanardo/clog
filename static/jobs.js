function updateScreen(jobs) {
	document.getElementById('jobs').innerHTML = jobs;
}

function loadJobs() {
	var req = XMLHttpRequest();
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
	window.setInterval(loadJobs, 10000);
}