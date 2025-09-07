FROM jenkins/jenkins:lts

USER root

# Install kubectl
RUN apt-get update && apt-get install -y curl gnupg lsb-release \
    && curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" \
    && install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl \
    && rm kubectl

# Install Docker CLI
RUN curl -fsSL https://get.docker.com/rootless | sh || true \
    && apt-get update && apt-get install -y docker.io \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

USER jenkins
