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
    locations.forEach(location => {
      const marker = markers.get(location.Slug);
      const hasStandard = filterStandards.size === 0 || filterStandards.has(location.Standard);
      const hasTags = filterTags.size === 0 || location.Tags.some(tag => filterTags.has(tag));

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

  function renderTagIcon(tag) {
    switch (tag) {
      case "beef":
        return `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6"><path stroke-linecap="round" stroke-linejoin="round" d="m21 7.5-9-5.25L3 7.5m18 0-9 5.25m9-5.25v9l-9 5.25M3 7.5l9 5.25M3 7.5v9l9 5.25m0-9v9" /></svg>`;
      case "pork":
        return `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6"><path stroke-linecap="round" stroke-linejoin="round" d="M12 6v12m-3-2.818.879.659c1.171.879 3.07.879 4.242 0 1.172-.879 1.172-2.303 0-3.182C13.536 12.219 12.768 12 12 12c-.725 0-1.45-.22-2.003-.659-1.106-.879-1.106-2.303 0-3.182s2.9-.879 4.006 0l.415.33M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>`;
      case "fish":
        return `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6"><path stroke-linecap="round" stroke-linejoin="round" d="M12 6v12m-3-2.818.879.659c1.171.879 3.07.879 4.242 0 1.172-.879 1.172-2.303 0-3.182C13.536 12.219 12.768 12 12 12c-.725 0-1.45-.22-2.003-.659-1.106-.879-1.106-2.303 0-3.182s2.9-.879 4.006 0l.415.33M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>`;
      case "produce":
        return `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6"><path stroke-linecap="round" stroke-linejoin="round" d="M14.121 7.629A3 3 0 0 0 9.017 9.43c-.023.212-.002.425.028.636l.506 3.541a4.5 4.5 0 0 1-.43 2.65L9 16.5l1.539-.513a2.25 2.25 0 0 1 1.422 0l.655.218a2.25 2.25 0 0 0 1.718-.122L15 15.75M8.25 12H12m9 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>`;
      case "poultry":
        return `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6"><path stroke-linecap="round" stroke-linejoin="round" d="M15 8.25H9m6 3H9m3 6-3-3h1.5a3 3 0 1 0 0-6M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>`;
      case "dairy":
        return `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6"><path stroke-linecap="round" stroke-linejoin="round" d="m9 7.5 3 4.5m0 0 3-4.5M12 12v5.25M15 12H9m6 3H9m12-3a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>`;
      case "grains":
        return `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6"><path stroke-linecap="round" stroke-linejoin="round" d="M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /><path stroke-linecap="round" stroke-linejoin="round" d="M15.91 11.672a.375.375 0 0 1 0 .656l-5.603 3.113a.375.375 0 0 1-.557-.328V8.887c0-.286.307-.466.557-.327l5.603 3.112Z" /></svg>`;
      case "shellfish":
        return `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v6m3-3H9m12 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>`;
      case "honey":
        return `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6"><path stroke-linecap="round" stroke-linejoin="round" d="M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /><path stroke-linecap="round" stroke-linejoin="round" d="M9 9.563C9 9.252 9.252 9 9.563 9h4.874c.311 0 .563.252.563.563v4.874c0 .311-.252.563-.563.563H9.564A.562.562 0 0 1 9 14.437V9.564Z" /></svg>`;
      case "wine":
        return `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6"><path stroke-linecap="round" stroke-linejoin="round" d="M9.879 7.519c1.171-1.025 3.071-1.025 4.242 0 1.172 1.025 1.172 2.687 0 3.712-.203.179-.43.326-.67.442-.745.361-1.45.999-1.45 1.827v.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 5.25h.008v.008H12v-.008Z" /></svg>`;
      case "beer":
        return `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6"><path stroke-linecap="round" stroke-linejoin="round" d="M15.182 15.182a4.5 4.5 0 0 1-6.364 0M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0ZM9.75 9.75c0 .414-.168.75-.375.75S9 10.164 9 9.75 9.168 9 9.375 9s.375.336.375.75Zm-.375 0h.008v.015h-.008V9.75Zm5.625 0c0 .414-.168.75-.375.75s-.375-.336-.375-.75.168-.75.375-.75.375.336.375.75Zm-.375 0h.008v.015h-.008V9.75Z" /></svg>`;
      case "patagonia":
        return `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6"><path stroke-linecap="round" stroke-linejoin="round" d="m6.115 5.19.319 1.913A6 6 0 0 0 8.11 10.36L9.75 12l-.387.775c-.217.433-.132.956.21 1.298l1.348 1.348c.21.21.329.497.329.795v1.089c0 .426.24.815.622 1.006l.153.076c.433.217.956.132 1.298-.21l.723-.723a8.7 8.7 0 0 0 2.288-4.042 1.087 1.087 0 0 0-.358-1.099l-1.33-1.108c-.251-.21-.582-.299-.905-.245l-1.17.195a1.125 1.125 0 0 1-.98-.314l-.295-.295a1.125 1.125 0 0 1 0-1.591l.13-.132a1.125 1.125 0 0 1 1.3-.21l.603.302a.809.809 0 0 0 1.086-1.086L14.25 7.5l1.256-.837a4.5 4.5 0 0 0 1.528-1.732l.146-.292M6.115 5.19A9 9 0 1 0 17.18 4.64M6.115 5.19A8.965 8.965 0 0 1 12 3c1.929 0 3.716.607 5.18 1.64" /></svg>`;
    }
  }

  function addMapLocations() {
    locations.forEach(location => {
      const isPatagonia = location.Tags.includes('patagonia') ? 'patagonia' : '';
      const marker = new mapboxgl.Marker()
        .setLngLat([location.Lat, location.Lng])
        .setPopup(
          new mapboxgl.Popup().setHTML(`
          <div class="location-popup flex flex-col ${
            location.Standard
          } ${isPatagonia}">
            ${
              location.Image
                ? `<div class="location-popup-image" style="background-image: url(\'/public/images/${location.Slug}-popup.jpg\')"></div>`
                : ""
            }

            <div class="location-popup-content">
              <ul class="tags">
                ${location.Tags.map(
                  (tag) => `<li class="tag">${renderTagIcon(tag)}</li>`
                ).join("")}
              </ul>
              <h3>${location.Name}</h3>
              <p>${location.Description}</p>
              <a class="outline-none button button-outline" href="/locations/${
                location.Slug
              }" target="_blank">Explore</a>
            </div>
          </div>
        `)
        );

      markers.set(location.Slug, marker);
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