# Nobl9 Project Bot User Guide

## Overview

The Nobl9 Project Bot is a Slack bot that helps you manage Nobl9 projects and user roles. It provides an interactive interface for creating projects, assigning roles, and managing project settings.

## Getting Started

### Prerequisites

- A Nobl9 account with API access
- A Slack workspace where you want to use the bot
- Required environment variables:
  - `NOBL9_API_KEY`: Your Nobl9 API key
  - `NOBL9_ORG`: Your Nobl9 organization name
  - `NOBL9_BASE_URL`: Nobl9 API base URL (default: https://api.nobl9.com)

### Installation

1. Add the bot to your Slack workspace
2. Configure the required environment variables
3. Start the bot service

## Commands

### Project Management

#### Create Project
```
/create
```
Starts an interactive project creation flow:
1. Enter project name
2. Provide project description
3. Confirm creation

#### List Projects
```
/list
```
Lists all projects in your Nobl9 organization.

### Role Management

#### Assign Role
```
/assign <project-name>
```
Starts an interactive role assignment flow:
1. Enter user email
2. Select role type (admin, member, viewer)
3. Confirm assignment

#### List Roles
```
/roles <project-name>
```
Lists all users and their roles in the specified project.

### General Commands

#### Help
```
/help
```
Shows available commands and their usage.

## Interactive Features

### Project Creation Flow

1. Start with `/create`
2. Enter a project name
   - Must be unique
   - Can contain letters, numbers, and hyphens
   - Cannot contain spaces or special characters
3. Provide a project description
   - Optional
   - Can be multiple lines
4. Confirm creation
   - Type `yes` to create
   - Type `no` to cancel

### Role Assignment Flow

1. Start with `/assign <project-name>`
2. Enter user email
   - Must be a valid email address
   - User must exist in Nobl9
3. Select role type
   - `admin`: Full project access
   - `member`: Standard project access
   - `viewer`: Read-only access
4. Confirm assignment
   - Type `yes` to assign
   - Type `no` to cancel

## Error Handling

### Common Errors

1. Project Name Already Exists
   - Error: "Project name is not available"
   - Solution: Choose a different project name

2. Invalid User
   - Error: "User not found"
   - Solution: Verify the email address

3. Rate Limit Exceeded
   - Error: "Rate limit exceeded"
   - Solution: Wait a few seconds and try again

4. Invalid Command
   - Error: "Unknown command"
   - Solution: Use `/help` to see available commands

### Recovery

- The bot automatically retries failed operations
- Rate-limited operations are retried with exponential backoff
- Failed operations can be cancelled with `no`

## Best Practices

1. Project Naming
   - Use descriptive names
   - Follow a consistent naming convention
   - Avoid special characters

2. Role Assignment
   - Assign minimum required permissions
   - Review role assignments regularly
   - Use project-specific roles when possible

3. Error Prevention
   - Verify project names before creation
   - Check user emails before assignment
   - Use the help command when unsure

## Troubleshooting

### Common Issues

1. Bot Not Responding
   - Check if the bot is online
   - Verify environment variables
   - Check API connectivity

2. Command Not Working
   - Verify command syntax
   - Check user permissions
   - Look for error messages

3. Rate Limiting
   - Reduce request frequency
   - Use batch operations when possible
   - Implement proper error handling

### Getting Help

- Use `/help` for command reference
- Check error messages for specific issues
- Contact support for persistent problems

## Security

### API Key Management

- Store API keys securely
- Rotate keys regularly
- Use environment variables

### Access Control

- Review role assignments
- Remove unused access
- Monitor user activity

## Support

For additional help:
- Use `/help` for command reference
- Check the documentation
- Contact Nobl9 support

## Updates

The bot is regularly updated with new features and improvements. Check the changelog for the latest updates. 