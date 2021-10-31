import { Sim } from '/tbc/core/sim.js';
import { TypedEvent } from '/tbc/core/typed_event.js';

import { Component } from './component.js';

declare var tippy: any;

/**
 * Data for creating a new input UI element.
 */
export interface InputConfig<ModObject, T> {
  label?: string,
	labelTooltip?: string,

  defaultValue?: T,

	// Returns the event indicating the mapped value has changed.
  changedEvent: (obj: ModObject) => TypedEvent<any>,

	// Get and set the mapped value.
  getValue: (obj: ModObject) => T,
  setValue: (obj: ModObject, newValue: T) => void,

	// If set, will automatically disable the input when this evaluates to false.
	enableWhen?: (obj: ModObject) => boolean,
}

// Shared logic for UI elements that are mapped to a value for some modifiable object.
export abstract class Input<ModObject, T> extends Component {
	private readonly inputConfig: InputConfig<ModObject, T>;
	readonly modObject: ModObject;

  constructor(parent: HTMLElement, cssClass: string, modObject: ModObject, config: InputConfig<ModObject, T>) {
    super(parent, 'input-root');
		this.inputConfig = config;
		this.modObject = modObject;
		this.rootElem.classList.add(cssClass);

    if (config.label) {
      const label = document.createElement('span');
      label.classList.add('input-label');
      label.textContent = config.label;
      this.rootElem.appendChild(label);

			if (config.labelTooltip) {
				tippy(label, {
					'content': config.labelTooltip,
					'allowHTML': true,
				});
			}
    }

    config.changedEvent(this.modObject).on(() => {
			this.setInputValue(config.getValue(this.modObject));
			this.update();
    });
	}

	private update() {
		const enable = !this.inputConfig.enableWhen || this.inputConfig.enableWhen(this.modObject);
		if (enable) {
			this.rootElem.classList.remove('disabled');
			this.getInputElem().removeAttribute('disabled');
		} else {
			this.rootElem.classList.add('disabled');
			this.getInputElem().setAttribute('disabled', '');
		}
	}

	// Can't call abstract functions in constructor, so need an init() call.
	init() {
		if (this.inputConfig.defaultValue) {
			this.setInputValue(this.inputConfig.defaultValue);
		} else {
			this.setInputValue(this.inputConfig.getValue(this.modObject));
		}
		this.update();
	}

	abstract getInputElem(): HTMLElement;

	abstract getInputValue(): T;

	abstract setInputValue(newValue: T): void;

	// Child classes should call this method when the value in the input element changes.
	inputChanged() {
		this.inputConfig.setValue(this.modObject, this.getInputValue());
	}
}
