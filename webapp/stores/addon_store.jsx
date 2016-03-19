/**
 * Created by enahum on 2/9/16.
 */

import AppDispatcher from 'dispatcher/app_dispatcher.jsx';
import EventEmitter from 'events';

import * as Utils from 'utils/utils.jsx';
import Constants from 'utils/constants.jsx';
const ActionTypes = Constants.ActionTypes;

const CHANGE_EVENT = 'change';

class AddonStoreClass extends EventEmitter {
    constructor() {
        super();

        this.emitChange = this.emitChange.bind(this);
        this.addChangeListener = this.addChangeListener.bind(this);
        this.removeChangeListener = this.removeChangeListener.bind(this);

        this.getAllAddons = this.getAllAddons.bind(this);
        this.getAddons = this.getAddons.bind(this);
        this.storeAddons = this.storeAddons.bind(this);
        this.updateAddon = this.updateAddon.bind(this);

        this.addons = {};
    }

    emitChange() {
        this.emit(CHANGE_EVENT);
    }

    addChangeListener(callback) {
        this.on(CHANGE_EVENT, callback);
    }

    removeChangeListener(callback) {
        this.removeListener(CHANGE_EVENT, callback);
    }

    getAllAddons() {
        return Object.keys(this.addons).map((key) => {
            return this.addons[key];
        });
    }

    getAddons(category) {
        const addons = this.getAllAddons();
        const array = addons.filter((c) => c.category === category);

        return Utils.sortByKey(array, 'name');
    }

    storeAddons(addons) {
        if (Array.isArray(addons)) {
            addons.forEach((a) => {
                this.addons[a.id] = a;
            });
        }
    }

    updateAddon(id, action) {
        switch (action) {
        case ActionTypes.ADDON_INSTALLED:
            this.addons[id].installed = true;
            break;
        case ActionTypes.ADDON_UNINSTALLED:
            this.addons[id].installed = false;
            break;
        }
    }
}

var AddonStore = new AddonStoreClass();

AddonStore.dispatchToken = AppDispatcher.register((payload) => {
    var action = payload.action;
    switch (action.type) {
    case ActionTypes.ADDON_INSTALLED:
    case ActionTypes.ADDON_UNINSTALLED: {
        AddonStore.updateAddon(action.addon_id, action.type);
        AddonStore.emitChange();
        break;
    }
    default:
    }
});

export default AddonStore;