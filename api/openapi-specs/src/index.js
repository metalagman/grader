import SwaggerUI from 'swagger-ui';
import StandalonePreset from 'swagger-ui/dist/swagger-ui-standalone-preset'
import 'swagger-ui/dist/swagger-ui.css';

const urls = require('./config-urls.yaml');

console.log(urls);

const ui = SwaggerUI({
  // spec,
  dom_id: '#swagger',
  urls: urls,
  deepLinking: true,
  docExpansion: 'list',
  presets: [
    SwaggerUI.presets.apis,
    StandalonePreset
  ],
  layout: "StandaloneLayout",
  tagsSorter: 'alpha'
});

// ui.initOAuth({
//   appName: "Swagger UI Webpack Demo",
//   // See https://demo.identityserver.io/ for configuration details.
//   clientId: 'implicit'
// });
