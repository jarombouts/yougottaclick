# Infinite Scroll Grid

This project implements an infinite scroll grid using Python and HTMX. The grid should dynamically load more content as the user scrolls down AND up, providing a seamless and efficient user experience.

## To do

This demo should have:

- **Responsive Grid Layout**: The grid layout adjusts based on the screen size.
- **Infinite Scrolling**: Automatically loads more content as the user scrolls.
- **Smooth Scrolling**: Ensures smooth transitions when new content is loaded.
- **Out-of-Band Swapping**: Efficiently manages DOM elements to keep the page performant.

## Technologies Used

- **Flask**: It's a dumb Python Flask app.
- **HTMX**: Handles AJAX requests and dynamic content loading.
- **HTML/CSS**: Frontend structure and styling.

It should be as simple and self contained as possible.

## File Structure

- `templates/index.html`: Main HTML file with the initial grid setup and JavaScript for handling scroll events.
- `templates/scroll.html`: Template for loading new grid batches.
- `templates/scroll-up.html`: Template for loading previous grid batches.
