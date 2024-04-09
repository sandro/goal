;(function() {
  'use strict';

  window.goal = window.goal || {url: "/"}

  const params = new URLSearchParams
  const loadedAt = new Date

  function isBot() {
    // Headless browsers are probably a bot.
    var w = window, d = document
    if (w.callPhantom || w._phantom || w.phantom)
      return true
    if (w.__nightmare)
      return true
    if (d.__selenium_unwrapped || d.__webdriver_evaluate || d.__driver_evaluate)
      return true
    if (navigator.webdriver)
      return true
    return false
  }

  function onLoad(callback) {
    document.addEventListener("visibilitychange", () => {
      if (document.visibilityState === "hidden") {
        callback()
      }
    })
  }

  function recordHit() {
    if (!window.goal.hit_id || isBot()) return
    params.set("hit_id", window.goal.hit_id)
    params.set("title", document.title)
    params.set("time_on_page", (new Date) - loadedAt)
    params.set("width", window.screen.width)
    params.set("height", window.screen.height)
    params.set("device_pixel_ratio", window.devicePixelRatio || 1)
    params.set("path", window.location.pathname)
    params.set("query", window.location.search)
    navigator.sendBeacon(window.goal.url, params)
  }
  onLoad(recordHit)
})();
