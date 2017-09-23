node {

	def cacheVolumeName = "${env.JOB_NAME}-${env.EXECUTOR_NUMBER}".replaceAll("/", "-");

	stage("prepare") {
		checkout scm
		sh "mkdir log"
		sh "docker build . --iidfile log/dockerid";
		sh "docker volume create ${cacheVolumeName}"
	}

	stage("build") {
		def dockerid = readFile("log/dockerid");
		sh "docker run --rm -v '${cacheVolumeName}:/cache' -v '${pwd()}/log:/log' '${dockerid}' --coverage"
	}

	stage("post") {
		archive includes: "log/**/*.log,log/**/*.html,log/**/*.out";
		publishHTML(target: [
			reportName: "Coverage",
			reportDir: "log",
			reportFiles: "coverage.html",
			keepAll: true,
			allowMissing: false,
			alwaysLinkToLastBuild: false,
		]);
	}
}
