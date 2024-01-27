# Boot.dev Web Servers Course

This is a repo of the code created while completing the "Learn Web Servers" course on Boot.dev. The course is a guided project on how to build a build a web server for a Twitter clone called Chirpy. Minimal implementation code was provided by the course (found primarily in the early part of the course), and the vast majority of the code is created by the person taking the course.

## Motivations

My professional experience up to this point didn't include web server code development or writing anything in Go, so I took the opportunity to learn a new facet of software development and test my understanding of a new language. You could technically use any language given the grading/testing of features was done by hitting the server endpoints rather than code submissions, but code samples and hints in Go made it easier to use Go instead of translating between languages while also trying to understand the concept in the task/assignment.

## Thoughts After Completion

Here are some thoughts I have after completing the course, or had during the course and never came up with a satisfactory resolution while working through the material:

- I think I prefer Go's error handling paradigm compared to a "standard" try-catch paradigm. Writing `if err != nil` a significant number of times does get annoying after a while though.
    - A side note on Go's error handling: I think my database tests weren't clear due to all of the error handling in the setup operations. Some refactoring is in order.
- Not knowing where the course was going ahead of time resulted in things as a single package to start, and things like all database or all API functionality being defined in single files. It became clearer where some refactoring could be done as more features/functionality was added. Examples include splitting the database and API functionality into different packages and splitting user, chirp, etc. related functionality into different files. This reinforced my current perception that, unless you've solved the same problem multiple times and have been able to refine a solution, development should start with getting _a_ reasonable solution to a feature request or problem and then it should be refined rather than trying to architect a solution before you have a good understanding of the problem space you're working in.
- Having never constructed a database before, predicting a good database structure was difficult. I have some better ideas after course completion:
    - Using a file as the database is clearly a problem for a real product, and the course noted it was a stopgap for not hooking up a real database to the server.
    - A real database could allow you to key users on both `id` and `email` to ensure neither are duplicated, which was a requirement for this course.
    - Splitting user data into something like a database/table with authentication information such as hashed passwords and auth tokens, and another database/table with the data requiring authentication such as account information (ex. subscription status) could increase security.
    - Splitting data into different databases/tables would allow for a bit of traffic management. I conceptually understand database sharding as a result of this course (from some reading external to the course), and it's obvious that a real system would need it for users and chirps.
- API tests were done via the Thunder Client extension in VSCode, and I did like having the tests in the same application as the code I was writing. I have used Postman before with organization collections, which was convenient for running the same requests as the other developers and effectively having them in source control. I don't have a strong preference at the moment, but I'd probably prefer Thunder Client if I figured out how to source control the requests.
- Passing the various API secrets into the API config factory function seems like a bad idea, but I'm not quite sure what the best practice is.

## Checkout/Install

If you wanted to run the code yourself, run `git clone https://github.com/trolfu/boot-dev-web-servers-course.git` to clone the repository.

You will need to add a `.env` file to the root module directory, which is ignored by git, and include the keys `JWT_SECRET` and `POLKA_API_KEY` in the form of `key=value`. `JWT_SECRET` is a key used to create and parse JWTs, and should be treated as a cryptographic secret. `POLKA_API_KEY` is an auth key provided in chapter 8 lesson 4 of the course on Boot.dev for a simulated webhook request, and is probably Boot.dev user specific. For the purposes of checking functionality, any request using the `/api/polka/webhooks` could pass `ApiKey <token>` in the authentication header, where `<token>` is the same value in the `.env` file.

Run `go build -o <fileName>` to build the server application.

Run `<fileName>` to run the server.

## Improvements/Additions

Below are additional tasks I have completed or intend to complete that are unrelated to the course requirements:

- [ ] Write documentation for the API endpoints
- [ ] Fix id assignment logic. Deletion breaks current id generation logic
- [ ] Add additional pages using HTMX
    - [ ] Login page
    - [ ] Add a user homepage. Probably requires cookies
    - [ ] Add user's chirps to homepage
- [ ] Configure an actual database for data
- [ ] Get API integration tests in source control
- [ ] Deploy server via a Docker container
- [ ] Add unit tests to CI
- [ ] Add API integration tests to CI