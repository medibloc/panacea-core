pipeline {
    agent any

    environment {
        APP_NAME = "panacea-core"
        GIT_COMMIT_HASH = sh (script: "git log -n 1 --pretty=format:'%H'", returnStdout: true)
	    IMAGE_NAME = "${APP_NAME}:${GIT_COMMIT_HASH}"
	    IMAGE_NAME_BUILD_ENV= "${IMAGE_NAME}-build-env"
	    ARTIFACT_DIR = "${WORKSPACE}/artifacts"
    }

    stages {
        stage('Start') {
            steps {
                slackSend (channel: '#alerts-ci', color: '#FFFF00', message: "STARTED: '${env.JOB_NAME} [${env.BUILD_NUMBER}]' (${env.BUILD_URL})")
                sh 'mkdir -p ${ARTIFACT_DIR}'
            }
        }
        stage('Build') {
            steps {
                echo 'Building..'
                sh 'docker build --target build-env -t ${IMAGE_NAME_BUILD_ENV} .'
                sh 'docker build -t ${IMAGE_NAME} .'
            }
        }
        stage('Test') {
            steps {
                echo 'Testing..'
                sh 'docker run --rm -v ${ARTIFACT_DIR}:/src/artifacts ${IMAGE_NAME_BUILD_ENV} make test'
            }
        }
        stage('Analyze') {
            steps {
                echo 'Analyzing..'
                sh 'docker run --rm ${IMAGE_NAME_BUILD_ENV} make get_tools lint'
            }
        }
        stage('Deploy') {
            steps {
                echo 'Deploying..'
            }
        }
        stage('Publish') {
            steps {
                echo 'Publishing..'
                publishHTML(target: [
                    allowMissing: false,
                    alwaysLinkToLastBuild: false,
                    keepAll: false,
                    reportDir: 'env.ARTIFACT_DIR',
                    reportFiles: 'coverage.html',
                    reportName: 'Code Coverage',
                    reportTitles: 'env.APP_NAME Code Coverage'
                ])
            }
        }
    }

    post {
        always {
            sh 'docker rmi ${IMAGE_NAME_BUILD_ENV} ${IMAGE_NAME} || true'
        }
        success {
            slackSend (channel: '#alerts-ci', color: '#00FF00', message: "SUCCESSFUL: '${env.JOB_NAME} [${env.BUILD_NUMBER}]' (${env.BUILD_URL})")
        }
        failure {
            slackSend (channel: '#alerts-ci', color: '#FF0000', message: "FAILED: '${env.JOB_NAME} [${env.BUILD_NUMBER}]' (${env.BUILD_URL})")
        }
    }
}
