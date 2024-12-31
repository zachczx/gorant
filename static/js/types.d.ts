import { HtmxHeaderSpecification, HtmxRequestConfig, HttpVerb } from 'htmx.org';

declare global {
	// interface requestConfig {
	// 	boosted: boolean;
	// 	useUrlParams: boolean;
	// 	formData: FormData;
	// 	parameters: Record<string, string>;
	// 	unfilteredFormData: FormData;
	// 	unfilteredParameters: Record<string, string>;
	// 	headers: HtmxHeaderSpecification;
	// 	target: HTMLElement;
	// 	verb: HttpVerb;
	// 	errors: HtmxElementValidationError[];
	// 	withCredentials: boolean;
	// 	timeout: number;
	// 	path: string;
	// 	triggeringEvent: Event;
	// }

	interface HtmxAfterRequest extends Event {
		detail: {
			elt: HTMLElement;
			xhr: XMLHttpRequest;
			target: HTMLElement;
			requestConfig: HtmxRequestConfig;
			successful: boolean;
			failed: boolean;
		};
	}

	interface HtmxAfterSwap extends Event {
		detail: {
			elt: HTMLElement;
			xhr: XMLHttpRequest;
			target: HTMLElement;
			requestConfig: HtmxRequestConfig;
		};
	}

	interface HtmxConfigRequest extends Event {
		detail: {
			parameters: Record<string, string>;
			unfilteredParameters: Record<string, string>;
			headers: HtmxHeaderSpecification;
			elt: HTMLElement;
			target: HTMLElement;
			verb: HttpVerb;
		};
	}
}

export {};
