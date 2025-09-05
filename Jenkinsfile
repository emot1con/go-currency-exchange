pipeline {
    agent any

    environment {
        goHome = tool 'myGo'
        PATH   = "${goHome}/bin:${env.PATH}"
    }

    stages {
        stage("Checkout") {
            steps {
                echo "=== Checkout Source Code ==="
                sh "go version"
                sh "echo PATH - $PATH"
                sh "echo BUILD_NUMBER - $env.BUILD_NUMBER"
            }
        }

        stage("Tests") {
            steps {
                parallel(
                    "Unit Tests": {
                        echo "=== Running Unit Tests ==="
                        sh "go test ./internal/service -v"
                    },
                    "Benchmark Tests": {
                        echo "=== Running Benchmark Tests ==="
                        sh "go test ./internal/service -bench=. -benchmem"
                    },
                    "Integration Tests": {
                        echo "=== Running Integration Tests ==="
                        // sh "INTEGRATION=1 go test -run TestIntegrationOnly -v"
                    },
                    "Coverage": {
                        echo "Running Code Coverage"
                        sh "go test ./internal/service -cover"
                    }
                )
            }
        }

        stage("Build Docker Image") {
            steps {
                script {
                    dockerImage = docker.build("numpyh/currency-exchange:${env.GIT_COMMIT}")
                }
            }
        }

        stage("Push Docker Image") {
            steps {
                script {
                    docker.withRegistry('', dockerhub) {
                        dockerImage.push()
                    }
                }
            }
        }
    }
}
