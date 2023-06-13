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
					var shouldRefresh = (history.builds.length != context.history.length || (history.builds.length > 0 && context.history.length > 0 && history.builds[0].status != context.history[0].status));
					context.history = history.builds;
					if (context.history.length > 0) {
						context.status = context.history[0].status;
						if (context.status == "inprogress") {
							context.showBuild(undefined, context.history[0].name);
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

			var action = appUrl + "/builds/" + context.projectName + "/";
			if (context.status == 'inprogress') {
				action += "interrupt/" + context.history[0].name
			} else {
				action += "build"
			}

			$.post(action, {
				success: data => {
					setTimeout(() => {context.loadHistory();}, 500);
				}
			});
		};

		context.showBuild = (event, buildNo) => {
			if (event != undefined) {
				event.preventDefault();
				event.stopPropagation();
			}

			$.get(appUrl + "/builds/" + context.projectName + "/" + buildNo, {
				success: data => {
					var build = JSON.parse(data);
					if (build.status == "inprogress") {
						setTimeout(() => {
							context.showBuild(undefined, buildNo);
						}, 1000);
					} else if (context.shownOutput != undefined && context.shownOutput.status == "inprogress" && build.status != "inprogress") {
						context.loadHistory();
					}
					context.shownOutput = build;
					context.refresh();
				}
			});
			
		};

		context.loadHistory();

		return context;
	});
}