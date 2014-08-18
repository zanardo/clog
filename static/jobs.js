var cur_screen = 'jobs';
var historyOffset = 0;

function updateScreen(jobs) {
	document.getElementById('jobstbl').innerHTML = jobs;
	document.getElementById('jobstbl').style.opacity = 1.0;
}

function loadJobs() {
	cur_screen = 'jobs';
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

function historyPrev() {
	historyOffset -= 1;
	if(historyOffset < 0) historyOffset = 0;
	loadHistory();
}

function historyNext() {
	historyOffset += 1;
	loadHistory();
}

function loadHistory() {
	cur_screen = 'history';
	document.getElementById('btnloadhistory').style.fontWeight = 'bold';
	document.getElementById('btnloadjobs').style.fontWeight = 'normal';
	document.getElementById('jobstbl').style.opacity = 0.5;
	window.location.hash = '#history';
	var req = new XMLHttpRequest();
	req.open("GET", "/history?offset="+historyOffset, true);
	req.onreadystatechange = function() {
		if(req.readyState == 4 && req.status == 200) {
			updateScreen(req.responseText);
		}
	}
	req.send(null);
}

function reload() {
	if(cur_screen == 'jobs') {
		loadJobs();
	}
	else if(cur_screen == 'history') {
		loadHistory();
	}
}

function load() {
	cur_screen = 'jobs';
	if(window.location.hash == '#history') {
		cur_screen = 'history';
		historyOffset = 0;
	}
	reload();
	window.setInterval(reload, 60000);
}
