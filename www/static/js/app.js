function initApp(name) {
	var app = ajsf(name, context => {
		context.projectName = name;

		context.hello = function() {
			console.log('hello');
		};

		context.status = "unknown";
		context.history = [];

		context.loadHistory = () => {
			$.get(appUrl + "/projects/" + context.projectName, {
				success: data => {
					var history = JSON.parse(data);
					var shouldRefresh = (history.jobs.length != context.history.length || (history.jobs.length > 0 && context.history.length > 0 && history.jobs[0].status != context.history[0].status));
					context.history = history.jobs;
					if (context.history.length > 0) {
						context.status = context.history[0].status;
						if (context.status == "inprogress") {
							context.showJob(undefined, context.history[0].name);
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
			} else {
				return "Start";
			}
		};

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
							context.showJob(undefined, jobNo);
						}, 1000);
					} else if (context.shownOutput != undefined && context.shownOutput.status == "inprogress" && job.status != "inprogress") {
						context.loadHistory();
					}
					context.shownOutput = job;
					context.refresh();
				}
			});
			
		};

		context.loadHistory();

		return context;
	});
}