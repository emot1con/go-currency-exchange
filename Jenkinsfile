pipeline {
    agent any

    // tools { myGo '1.23' }

    environment {
        goHome = tool 'myGo' // pastikan sudah ada konfigurasi Go di Jenkins
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

        stage("Unit Tests") {
            steps {
                echo "=== Running Unit Tests ==="
                sh "go test ./internal/service -v"
            }
        }

        stage("Benchmark Tests") {
            steps {
                echo "=== Running Benchmark Tests ==="
                sh "go test ./internal/service -bench=. -benchmem"
            }
        }

        stage("Integration Tests") {
            steps {
                echo "=== Running Integration Tests (requires server stopped) ==="
                echo "To run manually with a live server, use:"
                echo "INTEGRATION=1 go test -run TestIntegrationOnly -v"
            }
        }

        stage("Coverage") {
            steps {
                echo "=== Running Code Coverage ==="
                sh "go test ./internal/service -cover"
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
                    dockerImage = docker.build("numpyh/currency-exchange:${env.BUILD_ID}")
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

    post {
        always {
            echo "Cleaning up..."
        }
        success {
            echo "This build was successful!"
        }
        failure {
            echo "This build failed."
        }
    }
}
