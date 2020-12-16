export default {
  app: {
    login: '登录',
    logout: '注销',
    home: '主页',
    admin: '管理员',
    username: '用户名',
    groups: '所属组',
    file: '文件',
    folder: '文件夹',
    empty_list: '这里空空的',
    go_back: '返回',
    root_path: 'Root',
    toggle_to_list: '切换到列表模式',
    toggle_to_thumbnail: '切换到缩略图模式',
    toggle_sort: '切换排序方式',
    sort: {
      name_asc: '文件名升序',
      name_desc: '文件名降序',
      mod_time_asc: '修改时间升序',
      mod_time_desc: '修改时间降序',
      size_asc: '文件大小升序',
      size_desc: '文件大小降序'
    }
  },
  error: {
    not_allowed: '不允许的操作',
    not_found: '资源不存在',
    server_error: '服务器错误'
  },
  form: {
    required_msg: '{f}是必填的'
  },
  routes: {
    title: {
      users: '用户',
      groups: '用户组',
      drives: '盘',
      misc: '其他'
    }
  },
  md: {
    error: '渲染 Markdown 时出现错误'
  },
  dialog: {
    base: {
      ok: '确定'
    },
    open: {
      max_items: '最多可选择 {n} 个',
      n_selected: '已选 {n} 个'
    },
    text: {
      yes: '是',
      no: '否'
    },
    loading: {
      cancel: '取消'
    }
  },
  p: {
    admin: {
      oauth_connected: '已连接到 {p}',
      t_users: '用户',
      t_groups: '用户组',
      t_drives: '盘',
      t_misc: '其他',
      drive: {
        reload_drives: '重新加载盘',
        reload_tip: '编辑配置后，重新加载才可生效',
        name: '名称',
        type: '类型',
        operation: '操作',
        edit: '编辑',
        delete: '删除',
        add_drive: '添加盘',
        edit_drive: '编辑 {n}',
        save: '保存',
        cancel: '取消',
        configure: '配置',
        configured: '已配置',
        not_configured: '尚未配置',
        add: '添加',
        or_edit: ' 或编辑盘',
        f_name: '名称',
        f_enabled: '已启用',
        f_type: '类型',
        delete_drive: '删除盘',
        confirm_delete: '确认删除 {n}？'
      },
      user: {
        username: '用户名',
        operation: '操作',
        add_user: '添加用户',
        edit: '编辑',
        delete: '删除',
        edit_user: '编辑 {n}',
        groups: '所属组',
        save: '保存',
        cancel: '取消',
        add: '添加',
        or_edit: ' 或编辑用户',
        f_username: '用户名',
        f_password: '密码',
        delete_user: '删除用户',
        confirm_delete: '确认删除 {n}？'
      },
      group: {
        name: '名称',
        operation: '操作',
        add_group: '添加组',
        edit: '编辑',
        delete: '删除',
        edit_group: '编辑 {n}',
        users: '包含用户',
        save: '保存',
        cancel: '取消',
        add: '添加',
        or_edit: ' 或编辑组',
        f_name: '名称',
        delete_group: '删除组',
        confirm_delete: '确认删除 {n}？'
      },
      misc: {
        permission_of_root: '根路径权限',
        save: '保存',
        clean: '清除',
        clean_invalid: '清理无效的权限项/挂载项',
        clean_cache: '清除缓存',
        statistics: '统计信息',
        refresh_in: '{n} 秒后刷新',
        invalid_path_cleaned: '已清理 {n} 个无效的路径'
      },
      p_edit: {
        subject: '主体',
        rw: '(读/写)',
        policy: '策略',
        any: '任何',
        reject: '拒绝',
        accept: '接受'
      }
    },
    task: {
      empty: '现在没有任务',
      start: '开始',
      pause: '暂停',
      stop: '停止',
      remove: '移除',

      s_created: '已创建',
      s_starting: '开始',
      s_paused: '已暂停',
      s_stopped: '已停止',
      s_error: '错误',
      s_completed: '已完成'
    },
    home: {
      file_exists: '\'{n}\' 已存在，覆盖还是跳过？',
      apply_all: '记住选择',
      readme_loading: '加载 README...',
      readme_failed: '加载 README 失败',
      unsaved_confirm: '尚未保存，确认离开？'
    },
    new_entry: {
      new_item: '新建',
      upload_file: '上传文件',
      create_folder: '创建文件夹',
      upload_tasks: '上传任务',
      tasks_status: '上传 {p}',
      drop_tip: '拖放到这里以上传',
      invalid_folder_name: '无效的文件夹名称',
      confirm_stop_task: '确认停止该任务？',
      confirm_remove_task: '确认移除该任务，不可恢复？',
      file_exists: '文件已存在',
      file_exists_confirm: '\'{n}\' 已存在，覆盖还是跳过？',
      skip: '跳过',
      override: '覆盖'
    },
    login: {
      username: '用户名',
      password: '密码',
      login: '登录'
    }
  },
  hv: {
    download: {
      download: '下载'
    },
    permission: {
      save: '保存'
    },
    text_edit: {
      save: '保存'
    }
  },
  handler: {
    copy_move: {
      copy: '复制',
      move: '移动',
      copy_to: '复制到',
      move_to: '移动到',
      copy_desc: '复制文件',
      move_desc: '移动文件',
      copying: '正在复制 {n} {p}',
      moving: '正在移动 {n} {p}',
      copy_open_title: '选择复制到',
      move_open_title: '选择移动到',
      override_or_skip: '已存在时覆盖还是跳过？',
      override: '覆盖',
      skip: '跳过'
    },
    delete: {
      name: '删除',
      desc: '删除文件',
      confirm_n: '删除这 {n} 个文件？',
      confirm: '删除这个文件？',
      deleting: '正在删除 {n} {p}'
    },
    download: {
      name: '下载',
      desc: '下载文件'
    },
    image: {
      name: '画廊',
      desc: '查看图片'
    },
    media: {
      name: '播放',
      desc: '播放媒体'
    },
    mount: {
      name: '挂载到',
      desc: '将该项目挂载到其他位置',
      open_title: '选择挂载到'
    },
    permission: {
      name: '权限',
      desc: '设置该项目的权限'
    },
    rename: {
      name: '重命名',
      desc: '重命名这个项目',
      input_title: '重命名',
      invalid_filename: '无效的文件名'
    },
    text_edit: {
      edit_name: '编辑',
      view_name: '查看',
      edit_desc: '编辑这个文件',
      view_desc: '查看这个文件'
    }
  }
}
