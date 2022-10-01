[Home](../../README.md) / Git Usage

# Git Usage
## Branching
We follow a trunk based development workflow. All branches will be short live, regardless of naming convention or 
release strategy. This aims to minimise conflicts and wasted work. Learn more about  [Trunk Based Development](https://trunkbaseddevelopment.com/).

The different types of branches we may use are:

1. Feature branches - when working on new enhancements to the application or new features. 
2. Bug branches - when squashing a bug.

## Naming
All branches should have an issue in Github assigned to them, to guarantee this is the case, all branches must start with 
`GH-##-name-of-branch` where the `#` refers to the github issue number. For example, if I was working on a new feature called
Get Cards my branch might be called `GH-11-get-cards`

This may seem superfluous but assists github in autolinking and actually reduces the amount of work we need to do with git, trust me.
We use GIT issues to track past, current & future work, Github allows us to automatically link work to issues 
and maintain an audit trail, but it does require us to use Github in a particular way.

see [autolinked references and url](https://help.github.com/en/github/writing-on-github/autolinked-references-and-urls#issues-and-pull-requests) 
to learn more about how to reference issues in pull requests and other github.com features

## Pull Requests
Just as we link branches(and therefor commits) to an issue, linking a pull request to an issue is also a really good idea. 
You can link a pull request to an issue to show that a fix is in progress and to automatically close the issue when the pull request is merged.

You can link an issue to a pull request manually or using a supported keyword in the pull request description.

When you link a pull request to the issue the pull request addresses, collaborators can see that someone is working on the issue.

When you merge a linked pull request into the default branch of a repository, its linked issue is automatically closed.
learn more [here](https://help.github.com/en/github/managing-your-work-on-github/linking-a-pull-request-to-an-issue).

## Code Reviews
- [Code Review Guidelines](https://github.com/golang/go/wiki/CodeReviewComments)
- [Code Reviews](https://github.com/golang/go/wiki/CodeReview)