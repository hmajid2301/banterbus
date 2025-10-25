(function() {
  'use strict';

  let lastConnectionState = null;

  htmx.on('htmx:wsOpen', () => {
    if (lastConnectionState !== 'open') {
      window.toast('Connected', 'success');
      lastConnectionState = 'open';
    }
  });

  htmx.on('htmx:wsClose', () => {
    if (lastConnectionState !== 'closed') {
      window.toast('Disconnected - Reconnecting...', 'warning');
      lastConnectionState = 'closed';
    }
  });

  htmx.on('htmx:wsConnecting', () => {
    if (lastConnectionState !== 'connecting') {
      window.toast('Connecting...', 'info');
      lastConnectionState = 'connecting';
    }
  });

  htmx.on('htmx:wsError', () => {
    window.toast('Connection Error - Retrying...', 'failure');
    lastConnectionState = 'error';
  });
})();
