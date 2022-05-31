# minikube-forward-ports

Forwards service ports from a minikube instance running within a WSL environment
to localhost.

## Why does this exist?

Minikube creates a VM (within WSL?) where it hosts all its components (control plane,
pods, etc). Exposing a service on the minikube instance results in a service that's
only accessible within the WSL environment through the VM's IP. This app forwards
the exposed service to the hosting Windows environment, making it accessible at
localhost.

Also, I needed an excuse to have a go at Go.
[This](https://gist.github.com/kwalter94/bd5235da7a61bf84f6a0cbf7fde9cbe9) was my
initial idea but then I had to first manually run `minikube service <service-name>`
to expose the service within minikube's single machine cluster then run the script to
forward to Windows. This saves me typing out 1 (or more depending on how many ports
need to be exposed) additional command.

## Usage

Just run:

```sh
./minikube-forward-port name-of-service
```

Then access the service from your browser. The port that has been bound is logged
to the command line. Look for a line that looks something like
"Creating tunnel from http://192.168.49.2:32439 to http://localhost:32439." Find
the port on that line.
