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
                        # Clean up any existing container
                        docker stop currency-exchange-test || true
                        docker rm currency-exchange-test || true
                        
                        # Build a local image for testing
                        docker build -t currency-exchange-test:latest .
                        
                        # Run the container
                        docker run -d --name currency-exchange-test -p 8080:8080 currency-exchange-test:latest
                        
                        # Wait longer for the service to be ready
                        echo "Waiting for service to start..."
                        sleep 10
                        
                        # Check if container is running
                        docker ps | grep currency-exchange-test
                        
                        # Check container logs for any startup issues
                        echo "=== Container Logs ==="
                        docker logs currency-exchange-test
                        
                        # Test if the service is responding
                        echo "=== Testing Service Health ==="
                        curl -f http://localhost:8080/health || echo "Health check failed"
                        
                        # Run integration tests
                        INTEGRATION=1 go test -run TestIntegrationOnly -v
                        """
                    },
                    "Coverage": {
                        echo "Running Code Coverage"
                        sh "go test ./internal/service -cover"
                    }
                )
            }
            post {
                always {
                    sh """
                    # Clean up container regardless of test outcome
                    docker stop currency-exchange-test || true
                    docker rm currency-exchange-test || true
                    """
                }
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
