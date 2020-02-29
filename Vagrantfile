# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
    
  config.vm.box = "fedora/30-cloud-base"
  config.vm.box_version = "30.20190425.0"

  config.vm.define "agola-demo"
  config.vm.hostname = "agola-demo"
  
  config.vm.network "private_network", type: "dhcp"
  
  config.vm.provider "libvirt" do |lv|
    lv.qemu_use_session = false
    lv.memory = 2048
    lv.cpus = 2
  end

  config.vm.provider "virtualbox" do |vb|
    vb.customize ['modifyvm', :id, '--memory', '2048', '--cpus', '2']
  end

  config.vm.synced_folder ".", "/vagrant", disabled: true

  config.vm.provision "file", source: "./docker-compose.yml", destination: "docker-compose.yml"
  config.vm.provision "file", source: "./agola-setup", destination: "agola-setup"

  config.vm.provision "shell", privileged: false, inline: <<-'SHELL'
    ## get vm ip
    export IP_VM=$(ip addr show eth1 | grep "inet\b" | awk '{print $2}' | cut -d/ -f1)

    # install requirements
    sudo dnf config-manager --add-repo https://download.docker.com/linux/fedora/docker-ce.repo
    sudo dnf install -y --nobest docker-ce docker-ce-cli containerd.io docker-compose git vim-enhanced
    sudo usermod -G docker vagrant
    sudo systemctl enable docker --now
    sudo mkdir -p /data/agola /data/gitea /data/etcd && sudo chown vagrant:vagrant /data/agola /data/gitea /data/etcd
       
    # prepare user01 local environment
    sudo useradd user01
    echo password | sudo passwd --stdin user01
    sudo su - user01 -c 'ssh-keygen -t rsa -b 4096 -C "user01@example.com" -f /home/user01/.ssh/id_rsa -N ""'
    sudo su - user01 -c "echo -e \"Host ${IP_VM}\n\tStrictHostKeyChecking no\" > /home/user01/.ssh/config && chmod 600 /home/user01/.ssh/config"
    sudo su - user01 -c "git config --global user.email 'user01@example.com' && git config --global  user.name 'User01'"
 
    # create image for agola-setup
    sudo docker build -t agola-setup ./agola-setup
    # generate agola config file in /data/agola/config.yml
    sudo docker run --rm -v /data/agola:/data/agola agola-setup template ${IP_VM}

    # start containers
    sudo IP_VM=${IP_VM} docker-compose -p agola-demo up -d

    # wait for gitea is up & running
    until sudo docker logs gitea | grep Listen; do sleep 3; done
 
    # create user01 on gitea
    sudo docker exec gitea su git -c '/usr/local/bin/gitea admin create-user --username user01 --password password --email user01@example.com --admin'

    # setup gitea and agola services
    sudo docker run --rm -v /home/user01/.ssh/id_rsa.pub:/tmp/id_rsa.pub:ro agola-setup services ${IP_VM}

    # copy agola binary
    sudo docker cp agola:/bin/agola /usr/local/bin

    # setup local repo
    sudo su - user01 -c "git clone https://github.com/agola-io/agola-example-go.git"
    sudo su - user01 -c "cd agola-example-go && git remote remove origin && git remote add origin ssh://git@${IP_VM}:2222/user01/agola-example-go.git && git push -u origin master"
  
  SHELL

  config.vm.provision "shell", privileged: false, run: "always", inline: <<-'SHELL'
    ## get vm ip
    IP_VM=$(ip addr show eth1 | grep "inet\b" | awk '{print $2}' | cut -d/ -f1)
    
    # post-up message
    echo "#################################################################################"
    echo "#"
    echo "# AGOLA DEMO"
    echo "#"
    echo "# WEB access:"
    echo "#"
    echo "#     agola: http://${IP_VM}:8000      (user01/password)"
    echo "#     gitea: http://${IP_VM}:3000      (user01/password)"
    echo "#"
    echo "# SSH access with vagrant key:"
    echo "#"
    echo "#     [me@host] $ vagrant ssh"
    echo "#     [vagrant@agola-demo] $ sudo su - user01"
    echo "#     [user01@agola-demo] $ agola ..."
    echo "#"
    echo "# SSH access with user01 password:"
    echo "#"
    echo "#     [me@host] $ ssh user01@${IP_VM}    (password)"
    echo "#     [user01@agola-demo] $ agola ..."
    echo "#"
    echo "#################################################################################"

  SHELL
  
end
