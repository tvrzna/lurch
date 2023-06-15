function initApp(name) {
	var app = ajsf(name, context => {
		context.projectName = name;

		context.history = [];

		context.loadHistory = () => {
			$.get(appUrl + "/projects/" + context.projectName, {
				success: data => {
					var history = JSON.parse(data);
					var shouldRefresh = (history.jobs.length != context.history.length || (history.jobs.length > 0 && context.history.length > 0 && history.jobs[0].status != context.history[0].status));
					context.history = history.jobs;
					if (context.history.length > 0) {
						context.status = context.history[0].status;
						context.showJob(undefined, context.history[0].name);
						if (context.status == "inprogress") {
							context.setOutputCollapsed(false);
						}
					}
					if (shouldRefresh) {
						context.refresh();
					}
				}
			});
		};

		context.actionTitle = () => {
			if (context.status == "inprogress") {
				return "Interrupt";
			}
			return "Start";
		};

		context.actionClass = () => {
			if (context.status == "inprogress") {
				return "action action-interrupt";
			}
			return "action action-start";
		};

		context.statusValue = (job) => {
			if (job == undefined) {
				job = context.selectedJob;
			}
			if (job == undefined || job == undefined) {
				return "Not started"
			}
			switch (job.status) {
				case "finished":
					return "Finished";
				case "stopped":
					return "Stopped";
				case "failed":
					return "Failed";
				case "inprogress":
					return "Running";
				default:
					return "Unknown";
			}
		};

		context.startDateValue = (job) => {
			if (job == undefined) {
				job = context.selectedJob;
			}
			if (job == undefined || job.startDate == null) {
				return "Not started"
			}
			var d = new Date(job.startDate);

			return d.getFullYear() + "-" + context.padToTwo(d.getMonth()+1) + "-" + context.padToTwo(d.getDate()) +
				" " + context.padToTwo(d.getHours()) + ":" + context.padToTwo(d.getMinutes()) + ":" + context.padToTwo(d.getSeconds());
		};

		context.jobLength = (job) => {
			if (job == undefined) {
				job = context.selectedJob;
			}
			if (job == undefined || job.startDate == null) {
				return ""
			}
			var start = new Date(job.startDate);
			var end = new Date();
			if (job.endDate != null && job.status != "inprogress") {
				end = new Date(job.endDate);
			}

			return "(" + (Math.abs(start - end)/1000) + "s)";
		};

		context.padToTwo = (value) => {
			if (value != undefined && value.toString().length < 2) {
				return '0' + value;
			}
			return value;
		}

		context.performAction = (event) => {
			if (event != undefined) {
				event.preventDefault();
				event.stopPropagation();
			}

			var action = appUrl + "/jobs/" + context.projectName + "/";
			if (context.status == 'inprogress') {
				action += "interrupt/" + context.history[0].name
			} else {
				action += "start"
			}

			$.post(action, {
				success: data => {
					setTimeout(() => {context.loadHistory();}, 500);
				}
			});
		};

		context.setOutputCollapsed = (value) => {
			const className = 'collapsed';
			var title = $(context.rootElement).find('.job-title');
			var output = $(context.rootElement).find('.job-output');
			if (!value) {
				if (title.hasClass(className)) {
					title.removeClass(className);
				}
				if (output.hasClass(className)) {
					output.removeClass(className);
				}
			} else {
				if (!title.hasClass(className)) {
					title.addClass(className);
				}
				if (!output.hasClass(className)) {
					output.addClass(className);
				}
			}
		}

		context.toggleOutputCollapsed = (event) => {
			if (event != undefined) {
				event.preventDefault();
				event.stopPropagation();
			}

			context.setOutputCollapsed(!$(context.rootElement).find('.job-title').hasClass('collapsed'));
		}

		context.showJob = (event, jobNo) => {
			if (event != undefined) {
				event.preventDefault();
				event.stopPropagation();
			}

			$.get(appUrl + "/jobs/" + context.projectName + "/" + jobNo, {
				success: data => {
					var job = JSON.parse(data);
					if (job.status == "inprogress") {
						setTimeout(() => {
							context.showJob(event, jobNo);
						}, 1000);
					} else if (context.selectedJob != undefined && context.selectedJob.status == "inprogress" && job.status != "inprogress") {
						context.loadHistory();
					}
					context.selectedJob = job;
					if (event != undefined) {
						context.setOutputCollapsed(false);
					}
					var rootEl = $(context.rootElement);

					rootEl.attr('class', 'project job-status-' + job.status);
					context.refresh();
					rootEl.find('pre')[0].scrollTop = rootEl.find('pre')[0].scrollHeight;
				}
			});

		};

		context.isSelected = (name) => {
			if (context.selectedJob != undefined && context.selectedJob.name == name) {
				return " selected";
			}
			return "";
		};

		context.loadHistory();

		return context;
	});

	app.attribute('ajsf-style-class', (el, value, context) => {
		$(el).attr('class', value);
	});

	app.attribute('ajsf-title', (el, value, context) => {
		$(el).attr('title', value);
	});
}


$('[ajsf]').each((i, el) => {
	var appName = $(el).attr('ajsf');
	initApp(appName);
});