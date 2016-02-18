/**
 * Created by enahum on 2/9/16.
 */

import ModalStore from '../../stores/modal_store.jsx';
import AppDispatcher from '../../dispatcher/app_dispatcher.jsx';
import Constants from '../../utils/constants.jsx';
import * as Client from '../../utils/client.jsx';
import {FormattedMessage} from 'mm-intl';

var Modal = ReactBootstrap.Modal;
var ActionTypes = Constants.ActionTypes;

export default class InstallAddonModal extends React.Component {
    constructor(props) {
        super(props);

        this.handleInstall = this.handleInstall.bind(this);
        this.handleToggle = this.handleToggle.bind(this);
        this.handleHide = this.handleHide.bind(this);

        this.selectedList = null;

        this.state = {
            show: false,
            addon: null,
            error: ''
        };
    }

    componentDidMount() {
        ModalStore.addModalListener(ActionTypes.TOGGLE_ADDON_INSTALL_MODAL, this.handleToggle);
    }

    componentWillUnmount() {
        ModalStore.removeModalListener(ActionTypes.TOGGLE_ADDON_INSTALL_MODAL, this.handleToggle);
    }

    componentDidUpdate(prevProps, prevState) {
        if (this.state.show && !prevState.show) {
            setTimeout(() => {
                $(ReactDOM.findDOMNode(this.refs.installAddonBtn)).focus();
            }, 0);
        }
    }

    handleInstall() {
        const state = this.state;

        Client.installAddon(
            state.addon.id,
            () => {
                AppDispatcher.handleViewAction({
                    type: ActionTypes.ADDON_INSTALLED,
                    addon_id: state.addon.id
                });
                this.handleHide();
                if (state.addon.config_url) {
                    AppDispatcher.handleViewAction({
                        type: ActionTypes.TOGGLE_ADDON_CONFIG_MODAL,
                        value: true,
                        addon: state.addon
                    });
                }
            },
            (err) => {
                state.error = err.message;
                this.setState(state);
            }
        );
    }

    handleToggle(value, args) {
        this.setState({
            show: value,
            addon: args.addon,
            error: ''
        });
    }

    handleHide() {
        this.setState({show: false});
    }

    render() {
        if (!this.state.addon) {
            return null;
        }

        var error = null;
        if (this.state.error) {
            error = <div className='form-group has-error'><label className='control-label'>{this.state.error}</label></div>;
        }

        return (
            <Modal
                show={this.state.show}
                onHide={this.handleHide}
            >
                <Modal.Header closeButton={true}>
                    <Modal.Title>
                        <FormattedMessage
                            id='install_addon.confirm'
                            defaultMessage='Confirm the Install for {addon}'
                            values={{
                                addon: this.state.addon.name
                            }}
                        />
                    </Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <FormattedMessage
                        id='install_addon.question'
                        defaultMessage='Are you sure you want to install the addon {addon}?'
                        values={{
                            addon: this.state.addon.name
                        }}
                    />
                    <br />
                    <br />
                    {error}
                </Modal.Body>
                <Modal.Footer>
                    <button
                        type='button'
                        className='btn btn-default'
                        onClick={this.handleHide}
                    >
                        <FormattedMessage
                            id='install_addon.cancel'
                            defaultMessage='Cancel'
                        />
                    </button>
                    <button
                        ref='installAddonBtn'
                        type='button'
                        className='btn btn-primary'
                        onClick={this.handleInstall}
                    >
                        <FormattedMessage
                            id='install_addon.ok'
                            defaultMessage='Install'
                        />
                    </button>
                </Modal.Footer>
            </Modal>
        );
    }
}
