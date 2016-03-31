/**
 * Created by enahum on 3/24/16.
 */

import AppDispatcher from '../dispatcher/app_dispatcher.jsx';
import EventEmitter from 'events';
import Constants from 'utils/constants.jsx';

const ActionTypes = Constants.ActionTypes;

const CALL_EVENT = 'call';
const INCOMING_CALL_EVENT = 'incoming_call';
const CANCEL_CALL_EVENT = 'cancel_call';
const REJECTED_CALL_EVENT = 'rejected_call';
const CONNECT_CALL_EVENT = 'connect_call';
const NOT_SUPPORTED_EVENT = 'not_supported_call';
const CALL_FAILED_EVENT = 'call_failed';

class VideoCallStoreClass extends EventEmitter {
    emitCall(userId, isCaller) {
        this.emit(CALL_EVENT, userId, isCaller);
    }

    addCallListener(callback) {
        this.on(CALL_EVENT, callback);
    }

    removeCallListener(callback) {
        this.removeListener(CALL_EVENT, callback);
    }

    emitIncomingCall(incoming) {
        this.emit(INCOMING_CALL_EVENT, incoming);
    }

    addIncomingCallListener(callback) {
        this.on(INCOMING_CALL_EVENT, callback);
    }

    removeIncomingCallListener(callback) {
        this.removeListener(INCOMING_CALL_EVENT, callback);
    }

    emitCancelCall(call) {
        this.emit(CANCEL_CALL_EVENT, call);
    }

    addCancelCallListener(callback) {
        this.on(CANCEL_CALL_EVENT, callback);
    }

    removeCancelCallListener(callback) {
        this.removeListener(CANCEL_CALL_EVENT, callback);
    }

    emitRejectedCall() {
        this.emit(REJECTED_CALL_EVENT);
    }

    addRejectedCallListener(callback) {
        this.on(REJECTED_CALL_EVENT, callback);
    }

    removeRejectedCallListener(callback) {
        this.removeListener(REJECTED_CALL_EVENT, callback);
    }

    emitNotSupportedCall() {
        this.emit(NOT_SUPPORTED_EVENT);
    }

    addNotSupportedCallListener(callback) {
        this.on(NOT_SUPPORTED_EVENT, callback);
    }

    removeNotSupportedCallListener(callback) {
        this.removeListener(NOT_SUPPORTED_EVENT, callback);
    }

    emitConnectCall(call) {
        this.emit(CONNECT_CALL_EVENT, call);
    }

    addConnectCallListener(callback) {
        this.on(CONNECT_CALL_EVENT, callback);
    }

    removeConnectCallListener(callback) {
        this.removeListener(CONNECT_CALL_EVENT, callback);
    }

    emitFailedCall(call) {
        this.emit(CALL_FAILED_EVENT, call);
    }

    addFailedCallListener(callback) {
        this.on(CALL_FAILED_EVENT, callback);
    }

    removeFailedCallListener(callback) {
        this.removeListener(CALL_FAILED_EVENT, callback);
    }
}

var VideoCallStore = new VideoCallStoreClass();
VideoCallStore.setMaxListeners(0);

VideoCallStore.dispatchToken = AppDispatcher.register((payload) => {
    var action = payload.action;

    switch (action.type) {
    case ActionTypes.INITIALIZE_VIDEO_CALL:
        VideoCallStore.emitCall(action.user_id, true);
        break;
    case ActionTypes.INCOMING_VIDEO_CALL:
        VideoCallStore.emitIncomingCall(action.incoming);
        break;
    case ActionTypes.CANCEL_VIDEO_CALL:
        VideoCallStore.emitCancelCall(action.call);
        break;
    case ActionTypes.ANSWER_VIDEO_CALL:
        VideoCallStore.emitCall(action.user_id, false);
        break;
    case ActionTypes.REJECTED_VIDEO_CALL:
        VideoCallStore.emitRejectedCall();
        break;
    case ActionTypes.CONNECT_VIDEO_CALL:
        VideoCallStore.emitConnectCall(action.call);
        break;
    case ActionTypes.NOT_SUPPORTED_VIDEO_CALL:
        VideoCallStore.emitNotSupportedCall();
        break;
    case ActionTypes.FAILED_VIDEO_CALL:
        VideoCallStore.emitFailedCall(action.call);
        break;
    default:
    }
});

export default VideoCallStore;