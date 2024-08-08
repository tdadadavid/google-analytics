export interface TrackingData {
  type: "event" | "page";
  identity: string;
  userAgent: string;
  event: string;
  category: string;
  referrer: string;
  isTouchDevice: boolean;
}


export interface TrackingPayload {
  tracking: TrackingData;
  site_id: string;
}

export interface TrackerOptions {
  isTouchEnabled: boolean
}