import {  TrackerOptions, TrackingPayload } from "./interfaces"

class Tracker {

  private id: string = "";
  private siteId: string = "";
  private referrer = "";
  private isTouchDevice: boolean = false;

  constructor(siteId: string, ref: string, options: TrackerOptions) {
    this.siteId = siteId;
    this.referrer = ref;

    const customId = this.getSession("id");
    if (customId) {
      this.id = customId;
    }
   }

  track(event: string, category: "Page views" | any) {
    const payload: TrackingPayload = {
			tracking: {
				event,
				category,
				identity: this.id,
				referrer: this.referrer,
				userAgent: navigator.userAgent,
				isTouchDevice: this.isTouchDevice,
				type: category === "Page views" ? "page" : "event",
			},
			site_id: this.siteId,
    };
    console.log(payload);
    return this.trackReq(payload)
  }

  page(path: string) {
    this.track(path, "Page views");
  }

  identity(customId: string) {
    this.id = customId;
    this.setSession("id", this.id);
  }

  private getSession(key: string) {
    key = `__got_${key}__`;
    const s = sessionStorage.getItem(key);
    if (!s) return null;
    return JSON.parse(s);
  }

  private setSession(key: string, value: any) {
    key = `__got_${key}__`;
    sessionStorage.setItem(key, JSON.stringify(value));
  }

  private trackReq(payload: TrackingPayload) {
    const data = JSON.stringify(payload);
    const url = `http://localhost:7500/track?data=${btoa(data)}`;
    const img = new Image();
    img.src = url;
    console.log(img);
  }
}

((w: any, d) => {
  // this is to get the current site-id
  const ds = d.currentScript?.dataset;
  if (!ds) {
    console.error("you must have a data-siteid in your script tag");
    return;
  } else if (!ds.siteid) {
		console.error("you must have a data-siteid in your script tag");
		return;
	}

  // Track external referrals
  let externalRefs = "";
  const ref = ds.referrer;
  // if there is a referrer and the referrer is not the current site then add the referrer
  if (ref && ref.indexOf(`${w.location.protocol}://${w.location.host}`) == 0) {
    externalRefs = ref;
  }

  // Check if the current device accessing this site is touchscreen enabled.
  let isTouchEnabled = false;
  if (navigator.maxTouchPoints > 0 || "ontouchstart" in window) {
    isTouchEnabled = true;
  }

  const path = w.location.pathname;

  let tracker = new Tracker(ds.siteid, externalRefs, { isTouchEnabled });
  w._got = w._got || tracker;

  tracker.page(path);


  const history = window.history;
  if (history['pushState']) {
    const originalFunc = history["pushState"];
		history.pushState = function () {
			originalFunc.apply(this, arguments);
			tracker.page(w.location.pathname);
		};
		window.addEventListener("popstate", () => {
			tracker.page(w.location.pathname);
		});
  }
  w.addEventListener("hashchange", () => {
    tracker.page(w.location.pathname);
  }, false)

})(window, document)