# TFC Run Task Demo
This project is a demo of a [Terraform Cloud/Enterprise Pre-Plan RunTask](https://developer.hashicorp.com/terraform/cloud-docs/integrations/run-tasks#integration-details), that processes jobs from a job queue. It performs various operations on the job payload, such as downloading a configuration version, running regular expressions on a folder, and sending the task result to the callback URL.

## How to use
You can use this repo as a starting point for your own Run Tasks, but please be aware this is not "production ready". For example, there's no TLS support in the listener, so you would need to use it in something like a service mesh to ensure traffic is secured.

## Installation
To run this project, follow these steps:

Clone the repository:
``` bash
   git clone https://github.com/gitrgoliveira/tfc-run-task-demo.git
```

Build the project:
```bash
    make build
```

Run the executable:
```bash
   ./tfc-run-task-demo
```

## What it does
This project processes jobs from a job queue. The `processJobs` function continuously checks the job queue for new jobs and performs the necessary operations on each job payload.

To use this project, make sure to replace the placeholder logic in the `processJobs` function with your own job processing logic. You can modify the code to perform different operations based on your requirements.

## Configuration
The project requires a `patternsFile.txt` file in the same directory as the executable. This file contains the regular expression patterns to be used for matching in the `runRegexOnFolder` function. Make sure to populate this file with the desired patterns before running the project.

## Final Notes
Most of this code (99%?), including this README, was built with ChatGPT and Codeium VSCode extension.