function initApp(name) {
	var app = ajsf(name, context => {
		context.projectName = name;

		context.hello = function() {
			console.log('hello');
		};

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

		context.statusValue = () => {
			if (context.selectedJob == undefined || context.selectedJob.status == undefined) {
				return "Not started"
			}
			switch (context.selectedJob.status) {//"unknown", "finished", "stopped", "failed", "inprogress"
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

		context.startDateValue = () => {
			if (context.selectedJob == undefined || context.selectedJob.startDate == null) {
				return "Not started"
			}
			var d = new Date(context.selectedJob.startDate);

			return d.getFullYear() + "-" + context.padToTwo(d.getMonth()+1) + "-" + context.padToTwo(d.getDate()) +
				" " + context.padToTwo(d.getHours()) + ":" + context.padToTwo(d.getMinutes()) + ":" + context.padToTwo(d.getSeconds());
		};

		context.jobLength = () => {
			if (context.selectedJob == undefined || context.selectedJob.startDate == null) {
				return ""
			}
			var start = new Date(context.selectedJob.startDate);
			var end = new Date();
			if (context.selectedJob.endDate != null && context.selectedJob.status != "inprogress") {
				end = new Date(context.selectedJob.endDate);
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
					context.refresh();
					$(context.rootElement).find('pre')[0].scrollTop = $(context.rootElement).find('pre')[0].scrollHeight
				}
			});

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