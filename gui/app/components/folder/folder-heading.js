// Copyright 2016 Documize Inc. <legal@documize.com>. All rights reserved.
//
// This software (Documize Community Edition) is licensed under
// GNU AGPL v3 http://www.gnu.org/licenses/agpl-3.0.en.html
//
// You can operate outside the AGPL restrictions by purchasing
// Documize Enterprise Edition and obtaining a commercial license
// by contacting <sales@documize.com>.
//
// https://documize.com

import Ember from 'ember';
import NotifierMixin from '../../mixins/notifier';
import TooltipMixin from '../../mixins/tooltip';

const {
	computed,
} = Ember;

export default Ember.Component.extend(NotifierMixin, TooltipMixin, {
	folderService: Ember.inject.service('folder'),
	folderName: '',
	hasNameError: computed.empty('folderName'),
	editMode: false,
    isEditor: false,

	keyUp(e) {
		if (e.keyCode === 27) { // escape key
			this.send('onCancel');
		}
	},

	actions: {
		toggleEdit() {
			this.set('folderName', this.get('folder.name'));
			this.set('editMode', true);

			Ember.run.schedule('afterRender', () => {
				$('#folder-name').select();
			});
		},

		onSave() {
			if (this.get('hasNameError')) {
				return;
			}

			this.set('folder.name', this.get('folderName'));

			this.get('folderService').save(this.get('folder'));
			this.showNotification('Saved');

			this.set('editMode', false);
		},

		onCancel() {
			this.set('editMode', false);
		}
	}
});
