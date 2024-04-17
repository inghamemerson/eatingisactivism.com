const eia = (function() {
  const LOADED_SCRIPTS = new Set();
  const LOADED_STYLES = new Set
  let locations = null;
  let mapboxToken = null;
  let Map = null;
  let debugMode = false;

  function setLocations(locs) {
    locations = locs;
  }

  function setMapboxToken(token) {
    mapboxToken = token;
  }

  function waitForLibrary(lib, callback, timeout) {
    if (window[lib]) {
      if (debugMode) {
        console.debug(`${lib} is available`);
      }

      callback();
    } else {
      console.warn(`${lib} is not available yet, waiting...`);
      setTimeout(() => {
        waitForLibrary(lib, callback);
      }, timeout);
    }
  }

  function documentReady(fn) {
    document.addEventListener("DOMContentLoaded", () => {
      if (
        document.readyState === "interactive" ||
        document.readyState === "complete"
      ) {
        if (debugMode) {
          console.debug("Document is ready");
        }
        fn();
      }
    });
  }

  function isValidURL(url) {
    try {
      new URL(url);
      return true;
    } catch (e) {
      return false;
    }
  }

  function injectJS(url) {
    if (!isValidURL(url)) {
      console.error("Invalid URL:", url);
      return;
    }

    if (debugMode) {
      console.debug("Injecting library:", url);
    }

    if (LOADED_SCRIPTS.has(url)) {
      console.warn("Library already loaded, skipping:", url);
      return;
    }

    const script = document.createElement("script");
    script.type = "text/javascript";
    script.src = url;
    document.head.appendChild(script);

    LOADED_SCRIPTS.add(url);
  }

  function injectCSS(url) {
    if (!isValidURL(url)) {
      console.error("Invalid URL:", url);
      return;
    }

    if (debugMode) {
      console.debug("Injecting CSS:", url);
    }

    if (LOADED_STYLES.has(url)) {
      console.warn("CSS already loaded, skipping:", url);
      return;
    }

    const link = document.createElement("link");
    link.rel = "stylesheet";
    link.href = url;
    document.head.appendChild(link);

    LOADED_STYLES.add(url);
  }

  function addMapLocations() {
    locations.forEach(location => {
      new mapboxgl.Marker()
        .setLngLat([location.Lat, location.Lng])
        .setPopup(new mapboxgl.Popup().setHTML(`
          <div class="location-popup flex flex-col ${location.Standard}">
            ${location.Image ? `<div class="location-popup-image" style="background-image: url(\'/public/images/${location.Slug}-popup.jpg\')"></div>` : ''}

            <div class="location-popup-content">
              <ul class="tags">
                ${location.Tags.map(tag => `<li class="tag">${tag}</li>`).join('')}
              </ul>
              <ul class="badges">
                ${location.Badges.map(badge => `<li class="badge">${badge}</li>`).join('')}
              </ul>
              <h3>${location.Name}</h3>
              <p>${location.Description}</p>
              <a class="outline-none button button-outline" href="/locations/${location.Slug}" target="_blank">Explore</a>
            </div>
          </div>
        `))
        .addTo(Map);
    });
  }

  function renderHomeMap() {
    mapboxgl.accessToken = mapboxToken;

    Map = new mapboxgl.Map({
      attributionControl: false,
      compact: true,
      container: "map",
      style: "mapbox://styles/mapbox/outdoors-v12",
      center: [-98.5556199, 39.8097343],
      zoom: 2,
      minZoom: 2,
      maxZoom: 12,
      cooperativeGestures: true,
    });

    Map.on("load", () => {
      addMapLocations();
    });
  }

  function initMapbox() {
    const jsURL = "https://api.mapbox.com/mapbox-gl-js/v3.3.0/mapbox-gl.js";
    const cssURL = "https://api.mapbox.com/mapbox-gl-js/v3.3.0/mapbox-gl.css";

    // only run this if the #map element is present
    if (!document.getElementById("map")) {
      return;
    }

    // if we do not have a mapbox token, do not proceed
    if (!mapboxToken) {
      console.error("Mapbox token is missing");
      return;
    }

    // if the locations object is not present, do not proceed
    if (!locations) {
      console.error("Locations object is missing");
      return;
    }

    injectCSS(cssURL);
    injectJS(jsURL);

    waitForLibrary("mapboxgl", () => {
      renderHomeMap();
    }, 100);
  }

  function init(opts = {}) {
    if (opts.debug) {
      debugMode = true;
    }
    initMapbox();
  }

  return {
    init: init,
    setLocations: setLocations,
    setMapboxToken: setMapboxToken
  };
}());