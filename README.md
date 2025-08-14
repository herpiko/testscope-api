# Testscope API

<img src="https://github.com/herpiko/testscope-app/blob/main/public/screenshot.png?raw=true">

Testscope started as a small piece of software I built to help my teammates test what we were building—simple enough for anyone, even non-QA members, to use.

After four years, it went unmaintained, and I decided to fully abandon it due to the high running costs and the expense of the *.io domain. I assumed no one was still using it until the day it actually went offline, when people began reaching out. Unfortunately, the last database backup was from September 2024, and the VM was destroyed along with all the data.

I have since decided to open-source the code so anyone can self-host and manage it without relying on my service. I also plan to revive the service on a cheaper domain with a fair, sustainable subscription plan, so those without the capability to self-host can continue to use it.

This is my apology to those disappointed by the shutdown.

The code is a lot messy by my current standards, as I originally built it quickly solely to help my teammates. I am open to any suggestions and contributions. If you have any questions or feedback, feel free to reach out to me at herpiko@gmail.com.

This repository is the backend side of Testscope. The other related repositories are:

- [https://github.com/herpiko/testscope-app](https://github.com/herpiko/testscope-app) - frontend side
- [https://github.com/herpiko/testscope-deployment](https://github.com/herpiko/testscope-deployment) - simple deployment example

## Configuration

### Firebase

1. Setup your Firebase project and download the service account JSON configuration file.
2. Enable Google Sign-in method in Authentication.

This repository provided a working example of Firebase configuration but you need to setup your own.

### Growthbook

The backend system depends on Growthbook for feature flagging. Setup your Growthbook account and add this feature flag:

- `content_creation_limiter` and set it to `FALSE` as default.

### Env

1. Copy the `env.example` file to `.env` and adjust accordingly.
2. Adjust the `GROWTHBOOK_API_URL` value in the `.env` file
3. Adjust the `FIREBASE_PROJECT_ID` value in the `.env` file
4. Put your Firebase service account JSON file at `./firebase.json`.


## Preparation

- `make prep` - Spin up the infrastructure for local development.

This will also spawn a local proxy with SSL at [https://localhost:4443](https://localhost:4443).

## Run

- `make run` - Run the instance.

## Test
- `make test` - Run test
- `make cov` - Generate test coverage report

## Frontend Side

To work with frontend side locally, please clone [https://github.com/herpiko/testscope-app](https://github.com/herpiko/testscope-app) and follow the instructions in the `README.md` file.

Both frontend and backend supposed to be accessed at [https://localhost:4443](https://localhost:4443).

## License

This project is licensed under the AGPL license with additional terms - see the [LICENSE.md](LICENSE.md) file for details
