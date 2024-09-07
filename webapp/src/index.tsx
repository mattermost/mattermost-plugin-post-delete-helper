import {Action, Store} from 'redux';

import {Permissions} from 'mattermost-redux/constants';
import {getCurrentUser} from 'mattermost-redux/selectors/entities/common';
import {getConfig} from 'mattermost-redux/selectors/entities/general';
import {getPost} from 'mattermost-redux/selectors/entities/posts';
import {haveIChannelPermission} from 'mattermost-redux/selectors/entities/roles';
import {getCurrentTeam} from 'mattermost-redux/selectors/entities/teams';
import {GlobalState} from 'mattermost-redux/types/store';

import manifest from '@/manifest';

import {PluginRegistry} from '@/types/mattermost-webapp';

import {triggerDeleteRootPostCommand} from './actions';

export default class Plugin {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars, @typescript-eslint/no-empty-function
    public async initialize(registry: PluginRegistry, store: Store<GlobalState, Action<Record<string, unknown>>>) {
        // @see https://developers.mattermost.com/extend/plugins/webapp/reference/
        registry.registerPostDropdownMenuAction(
            'Remove root post',
            async (postID) => {
                store.dispatch(triggerDeleteRootPostCommand(postID) as any);
            },
            (postID) => {
                const state = store.getState();
                const post = getPost(state, postID);
                if (!post) {
                    return false;
                }

                console.debug('reply_count=', post.reply_count); //eslint-disable-line no-console

                // check if post has replies
                if (post.reply_count === 0) {
                    return false;
                }

                // Check if the user has permissions to edit his own post or edit other's posts if not the author
                const user = getCurrentUser(state);
                const team = getCurrentTeam(state);
                let permission = Permissions.EDIT_POST;
                if (post.user_id !== user.id) {
                    permission = Permissions.EDIT_OTHERS_POSTS;
                }
                if (!haveIChannelPermission(state, {
                    team: team.id,
                    channel: post.channel_id,
                    permission,
                })) {
                    return false;
                }

                // Check if post is editable
                const config = getConfig(state);
                const edit_time_limit : number = config.PostEditTimeLimit ? Number(config.PostEditTimeLimit) : -1;
                if (edit_time_limit !== -1 && post.create_at + (edit_time_limit * 1000) < Date.now()) {
                    return false;
                }

                return true;
            },
        );
    }
}

declare global {
    interface Window {
        registerPlugin(pluginId: string, plugin: Plugin): void
    }
}

window.registerPlugin(manifest.id, new Plugin());
