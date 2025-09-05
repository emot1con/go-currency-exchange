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
                        sh """
                        docker run -d --name test-app -p 8080:8080 my-app:latest
                        sleep 3
                        INTEGRATION=1 go test -run TestIntegrationOnly -v
                        docker stop test-app
                        docker rm test-app
                        """
                    },
                    "Coverage": {
                        echo "Running Code Coverage"
                        sh "go test ./internal/service -cover"
                    }
                )
            }
        }

        stage("Build Binary") {
            steps {
                echo "=== Building Go Application ==="
                sh "go build -o currency-exchange ./cmd"
            }
        }

        stage("Build Docker Image") {
            steps {
                script {
                    dockerImage = docker.build("numpyh/currency-exchange:${env.BUILD_TAG}")
                }
            }
        }

        stage("Push Docker Image") {
            steps {
                script {
                    docker.withRegistry('', 'dockerhub') {
                        dockerImage.push()
                        dockerImage.push('latest')
                    }
                }
            }
        }
    }
}
