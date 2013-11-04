var screen;

function updateScreen(jobs) {
	document.getElementById('jobstbl').innerHTML = jobs;
	document.getElementById('jobstbl').style.opacity = 1.0;
}

function loadJobs() {
	screen = 'jobs';
	document.getElementById('btnloadjobs').style.fontWeight = 'bold';
	document.getElementById('btnloadhistory').style.fontWeight = 'normal';
	document.getElementById('jobstbl').style.opacity = 0.5;
	window.location.hash = '#jobs';
	var req = new XMLHttpRequest();
	req.open("GET", "/jobs", true);
	req.onreadystatechange = function() {
		if(req.readyState == 4 && req.status == 200) {
			updateScreen(req.responseText);
		}
	}
	req.send(null);
}

function loadHistory() {
	screen = 'history';
	document.getElementById('btnloadhistory').style.fontWeight = 'bold';
	document.getElementById('btnloadjobs').style.fontWeight = 'normal';
	document.getElementById('jobstbl').style.opacity = 0.5;
	window.location.hash = '#history';
	var req = new XMLHttpRequest();
	req.open("GET", "/history", true);
	req.onreadystatechange = function() {
		if(req.readyState == 4 && req.status == 200) {
			updateScreen(req.responseText);
		}
	}
	req.send(null);
}

function reload() {
	if(screen == 'jobs') {
		loadJobs();
	}
	else if(screen == 'history') {
		loadHistory();
	}
}

function load() {
	screen = 'jobs';
	if(window.location.hash == '#history') {
		screen = 'history';
	}
	reload();
	window.setInterval(reload, 60000);
}