function initApp(name) {
	var app = ajsf(name, (context, rootEl) => {
		context.projectName = name;
		context.history = [];
		context.loading = 0;

		var loadingOverlay = $(rootEl).find('.loading-overlay');

		context.addLoading = (count) => {
			if (count == undefined) {
				count = 1;
			}
			context.loading += count;

			if (context.loading > 0) {
				setTimeout(() => {
					if (context.loading > 0) {
						loadingOverlay[0].style.display = 'flex';
					}
				}, 150)
			} else {
				loadingOverlay[0].style.display = 'none';
				context.loading = 0;
			}
		};

		context.showMessage = (type, message) => {
			var messageDiv = $("<div>" + message + "</div>")
				.attr('class', 'message ' + type)
				.appendTo($("#messageBox"));

			messageDiv.click(() => {
				messageDiv.remove();
			});

			setTimeout(() => {
				messageDiv.remove();
			}, 3000);
		};

		context.loadHistory = () => {
			context.addLoading();
			$.get(appUrl + "/projects/" + context.projectName, {
				success: data => {
					var detail = JSON.parse(data);
					var shouldRefresh = (detail.jobs.length != context.history.length || (detail.jobs.length > 0 && context.history.length > 0 && detail.jobs[0].status != context.history[0].status));
					context.history = detail.jobs;
					if (context.history.length > 0) {
						if (context.selectedJob == undefined) {
							var rootEl = $(context.rootElement);
							rootEl.attr('class', 'project job-status-' + context.history[0].status + (rootEl.hasClass('maximized') ? ' maximized': ''));
						}
						context.status = context.history[0].status;
						context.showJob(undefined, context.history[0].name);
						if (context.status == "inprogress") {
							context.setOutputCollapsed(false);
						}
					}

					context.removeAllParams();

					for (const key in detail.params) {
						context.addParamLine(undefined, key, detail.params[key]);
					}
					context.addParamLine();

					if (shouldRefresh) {
						context.refresh();
					}
				},
				error: () => {
					context.showMessage('error', 'Could not load ' + context.projectName);
				},
				complete: () => {
					context.addLoading(-1);
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

			var params = {};
			var action = appUrl + "/jobs/" + context.projectName + "/";
			var actionName = "start";
			if (context.status == 'inprogress') {
				action += "interrupt/" + context.history[0].name
				actionName = "interrupt";
			} else {
				action += "start"

				var configEl = $(context.rootElement).find('.project-config');
				if (!configEl.hasClass('collapsed')) {
					configEl.find('.param-line').each((i, el) => {
						var key = $(el).find('[name="key"]').val()
						var val = $(el).find('[name="value"]').val()

						if (key != '') {
							params[key] = val;
						}
					});
				}
			}

			context.addLoading();
			$.post(action, {
				data: {'params': params},
				success: data => {
					var msg = context.projectName + ' ' + actionName + 'ed';
					if (actionName == 'start' && Object.keys(params).length > 0) {
						msg += ' with parameters'
					}
					context.showMessage('info', msg);
					setTimeout(() => {context.loadHistory();}, 100);
				},
				error: () => {
					context.showMessage('error', 'Could not ' + actionName + ' ' + context.projectName);
				},
				complete: () => {
					context.addLoading(-1);
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

			let el = $(rootEl);
			if (el.hasClass('maximized')) {
				return;
			}

			context.setOutputCollapsed(!el.find('.job-title').hasClass('collapsed'));
		}

		context.maximize = (event) => {
			if (event != undefined) {
				event.preventDefault();
				event.stopPropagation();
			}

			let el = $(rootEl);
			if (el.hasClass('maximized')) {
				el.removeClass('maximized');
			} else {
				el.addClass('maximized');
			}
		}

		context.showJob = (event, jobNo, el, hideLoading) => {
			if (event != undefined) {
				event.preventDefault();
				event.stopPropagation();
			}

			if (hideLoading == undefined || !hideLoading) {
				context.addLoading();
			}
			$.get(appUrl + "/jobs/" + context.projectName + "/" + jobNo, {
				success: data => {
					var job = JSON.parse(data);
					if (job.status == "inprogress") {
						setTimeout(() => {
							context.showJob(event, jobNo, undefined, true);
						}, 1000);
					} else if (context.selectedJob != undefined && context.selectedJob.status == "inprogress" && job.status != "inprogress") {
						context.loadHistory();
					}
					context.selectedJob = job;
					if (event != undefined) {
						context.setOutputCollapsed(false);
					}
					var rootEl = $(context.rootElement);
					rootEl.attr('class', 'project job-status-' + job.status + (rootEl.hasClass('maximized') ? ' maximized' : ''));
					context.refresh();
					rootEl.find('pre')[0].scrollTop = rootEl.find('pre')[0].scrollHeight;
				},
				error: () => {
					context.showMessage('error', 'Could not load job #' + jobNo + ' from ' + context.projectName);
				},
				complete: () => {
					if (hideLoading == undefined || !hideLoading) {
						context.addLoading(-1);
					}
				}
			});

		};

		context.isSelected = (name) => {
			if (context.selectedJob != undefined && context.selectedJob.name == name) {
				return " selected";
			}
			return "";
		};

		context.artifactDownloadUrl = () => {
			if (context.selectedJob != undefined && context.selectedJob.name != undefined) {
				return appUrl.replace('rest', '') + "download/" + context.projectName + "/" + context.selectedJob.name;
			}
			else "";
		};

		context.artifactExists = () => {
			return context.selectedJob != undefined && context.selectedJob.artifactSize != undefined && context.selectedJob.artifactSize > 0;
		};

		context.artifactSize = () => {
			if (context.selectedJob != undefined && context.selectedJob.artifactSize != undefined && context.selectedJob.artifactSize > 0) {
				return context.selectedJob.artifactSize.toFixed(2) + ' ' + context.selectedJob.artifactUnit + 'B';
			}
			return "";
		}

		context.downloadArtifact = (event) => {
			if (event != undefined) {
				event.preventDefault();
				event.stopPropagation();
			}

			var link = $('<a href="' + context.artifactDownloadUrl() + '"></a>');
			link.click();
			link.remove();
		};

		context.toggleConfig = (event) => {
			if (event != undefined) {
				event.preventDefault();
				event.stopPropagation();
			}

			var configEl = $(context.rootElement).find('.project-config');
			if (configEl.hasClass('collapsed')) {
				configEl.removeClass('collapsed');
				context.fixConfigHeight(configEl);
			} else {
				configEl.addClass('collapsed');
				configEl.prop('style', 'height: 0');
			}
		};

		context.fixConfigHeight = (configEl) => {
			var totalHeight = 0;
			configEl.find(".project-config > *").each(function(i, el){
				totalHeight += el.offsetHeight;
			});

			if (!configEl.hasClass('collapsed')) {
				configEl.prop('style', 'height: ' + totalHeight + 'px;');
			}
		};

		context.addParamLine = (event, key, value) => {
			if (event != undefined) {
				event.preventDefault();
				event.stopPropagation();
			}

			var configEl = $(context.rootElement).find('.project-config');

			var paramLineEl = $('<div class="param-line"></div>').appendTo(configEl);
			var keyEl = $('<input type="text" name="key" placeholder="Parameter" />').appendTo(paramLineEl);
			var valueEl = $('<input type="text" name="value" placeholder="Value" />').appendTo(paramLineEl);
			var removeEl = $('<span class="remove-line"></span>').appendTo(paramLineEl);

			if (key != undefined && key != '') {
				keyEl.val(key);
				valueEl.val(value);
			}

			removeEl.click((event) => {
				if (event != undefined) {
					event.preventDefault();
					event.stopPropagation();
				}
				if ((keyEl.val() != undefined && keyEl.val() != '') || (valueEl.val() != undefined && valueEl.val() != '')) {
					paramLineEl.remove();
					context.fixConfigHeight(configEl);
				}
			});

			paramLineEl.find('input[type=text]').on('keyup', () => {
				var emptyCounter = 0;
				var paramLines = configEl.find(".param-line");
				for (var i = paramLines.length; i >= 0; i--) {
					var lineKey = $(paramLines[i]).find('input[name="key"]');
					var lineValue = $(paramLines[i]).find('input[name="value"]');
					if ((lineKey.val() == undefined || lineKey.val() == '') && (lineValue.val() == undefined || lineValue.val() == '')) {
						emptyCounter++;
						if (emptyCounter > 1) {
							$(paramLines[i]).remove();
							context.fixConfigHeight(configEl);
							continue;
						}
					}
				}
				if ((keyEl.val() != undefined && keyEl.val() != '') || (valueEl.val() != undefined && valueEl.val() != '')) {
					if (emptyCounter == 0) {
						context.addParamLine();
						context.fixConfigHeight(configEl);
					}
				}
			});
		};

		context.removeAllParams = () => {
			$(context.rootElement).find('.project-config .param-line').remove();
		};

		context.loadHistory();

		return context;
	});

	app.attribute('ajsf-style-class', (el, value) => {
		$(el).attr('class', value);
	});

	app.attribute('ajsf-title', (el, value) => {
		$(el).attr('title', value);
	});

	app.attribute('ajsf-href', (el, value) => {
		$(el).attr('href', value);
	});
}

$('[ajsf]').each((i, el) => {
	var appName = $(el).attr('ajsf');
	if (appName != "lurch-settings") {
		initApp(appName);
	}
});

ajsf("lurch-settings", context => {
	const themeItemName = "lurch.theme";

	context.switchTheme = () => {
		var currentTheme = localStorage.getItem(themeItemName);
		context.setTheme(currentTheme == "dark" ? "light" : "dark");
	};

	context.setTheme = (theme) => {
		if (theme == "dark") {
			$('body').attr("id", "dark");
		} else {
			$('body').removeAttr("id");
		}
		localStorage.setItem(themeItemName, theme);
	};

	context.setTheme(localStorage.getItem(themeItemName));
});