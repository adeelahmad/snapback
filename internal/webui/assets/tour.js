// tour.js: hides the first-run tour once the visitor dismisses it.
'use strict';

var TOUR_KEY = 'snapback.tour.dismissed';

function tourDismissed() {
  try {
    return localStorage.getItem(TOUR_KEY) === '1';
  } catch (err) {
    return false;
  }
}

function rememberTourDismissed() {
  try {
    localStorage.setItem(TOUR_KEY, '1');
  } catch (err) {
    // Storage may be blocked; hiding the tour for this page view is enough.
  }
}

function hideTour() {
  var parts = document.querySelectorAll('[data-tour]');
  for (var i = 0; i < parts.length; i++) {
    parts[i].hidden = true;
  }
}

function initTour() {
  if (tourDismissed()) {
    hideTour();
  }
  var buttons = document.querySelectorAll('[data-js="tour-dismiss"]');
  for (var i = 0; i < buttons.length; i++) {
    buttons[i].addEventListener('click', function () {
      rememberTourDismissed();
      hideTour();
    });
  }
}

document.addEventListener('DOMContentLoaded', initTour);
