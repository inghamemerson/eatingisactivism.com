const eia = (function() {
  const LOADED_SCRIPTS = new Set();
  const LOADED_STYLES = new Set();
  const markers = new Map();
  let locations = null;
  let mapboxToken = null;
  let Mapbox = null;
  let debugMode = false;
  let filterStandards = new Set();
  let filterTags = new Set();

  function setLocations(locs) {
    locations = locs;
  }

  function setMapboxToken(token) {
    mapboxToken = token;
  }

  function filterLocations() {
    if (debugMode) {
      console.debug("Filtering locations...");
      console.debug("Filter standards:", filterStandards);
      console.debug("Filter tags:", filterTags);
    }

    Object.keys(locations).forEach(slug => {
      const location = locations[slug];
      if (debugMode) {
        console.debug("Filtering location:", location.slug);
        console.log(location)
      }
      const marker = markers.get(location.slug);
      const hasStandard = filterStandards.size === 0 || filterStandards.has(location.standard.slug);
      const hasTags = filterTags.size === 0 || location.tags.some(tag => filterTags.has(tag.slug));

      if (hasStandard && hasTags) {
        marker.addTo(Mapbox);
      } else {
        marker.remove();
      }
    });
  }

  function waitForLibrary(lib, callback, timeout) {
    if (window[lib]) {
      if (debugMode) {
        console.debug(`${lib} is available`);
      }

      callback();
    } else {
      if (debugMode) {
        console.warn(`${lib} is not available yet, waiting...`);
      }
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
      if (debugMode) {
        console.error("Invalid URL:", url);
      }
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
      if (debugMode) {
        console.error("Invalid URL:", url);
      }
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
    Object.keys(locations).forEach(slug => {
      const location = locations[slug];
      const el = document.createElement("div");
      const isPatagonia = location.tags.includes('patagonia') ? 'patagonia-provisions' : '';
      el.className = `marker ${ location.standard } ${ isPatagonia }`;
      const marker = new mapboxgl.Marker(el)
        .setLngLat([location.lng, location.lat])
        .setPopup(
          new mapboxgl.Popup().setHTML(`
          <div class="location-popup flex flex-col ${
            location.standard
          } ${isPatagonia}">
            <div class="location-popup-content">
              <ul class="tags">
                ${location.tags.map(
                  (tag) => `<li class="tag">${tag.icon}</li>`
                ).join("")}
              </ul>
              <h3>${location.name}</h3>
              <p>${location.shortDescription}</p>
              <a class="outline-none button button-outline" href="/locations/${
                location.slug
              }" target="_blank">Explore</a>
            </div>
          </div>
        `)
        );

      markers.set(location.slug, marker);
    });

    filterLocations();
  }

  function renderHomeMap() {
    mapboxgl.accessToken = mapboxToken;

    Mapbox = new mapboxgl.Map({
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

    Mapbox.on("load", () => {
      addMapLocations();
      initFilterListeners();
      initFilterToggle();
    });
  }

  function initFilterToggle() {
    const filterToggle = document.getElementById("filterToggle");
    const filterPanel = document.getElementById("mapFilters");
    let mapPadding = {
      left: 0
    };

    if (!filterToggle || !filterPanel) {
      return;
    }

    filterToggle.addEventListener("click", () => {
      filterPanel.classList.toggle("show");

      if (filterPanel.classList.contains("show")) {
        mapPadding.left = 300;
      } else {
        mapPadding.left = 0;
      }

      Mapbox.easeTo({
        padding: mapPadding,
        duration: 240,
      });
    });
  }

  function initFilterListeners() {
    const standardFilters = document.querySelectorAll("#filter-standards .checkbox input");
    const tagFilters = document.querySelectorAll("#filter-tags .checkbox input");

    standardFilters.forEach(filter => {
      filterStandards.add(filter.value);
      filter.addEventListener("change", function() {
        const standard = this.value;
        const isChecked = this.checked;

        if (isChecked) {
          filterStandards.add(standard);
        } else {
          filterStandards.delete(standard);
        }

        filterLocations();
      });
    });

    tagFilters.forEach(filter => {
      filterTags.add(filter.value);
      filter.addEventListener("change", function() {
        const tag = this.value;
        const isChecked = this.checked;

        if (isChecked) {
          filterTags.add(tag);
        } else {
          filterTags.delete(tag);
        }

        filterLocations();
      });
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
    }, 200);
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