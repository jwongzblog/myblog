import sys
import random
import subprocess

split_factors = [["0sAQAAAAA=", "0sAQMAAAA=", "0sAQYAAAA="],
                 ["0sAQkAAAA=", "0sAQwAAAA=", "0sAQ8AAAA="],
                 ["0sARIAAAA=", "0sAQEAAAA=", "0sAQgAAAA="]]

class OSDMap(object):
    def __init__(self):
        self.server1 = [0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17]
        self.server2 = [22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39]
        self.server3 = [44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61]

    def find_server(self, osd_id):
        if osd_id in self.server1:
            return 'openstack-dev-1'
        if osd_id in self.server2:
            return 'openstack-dev-2'
        if osd_id in self.server3:
            return 'openstack-dev-3'
        else:
            return None

def execute_cmd(cmd):
    print cmd
    return subprocess.check_output(cmd.split()).split('\n')

def op_split_factor(op, osd_id, pg_name, split_factor):
    osdMap = OSDMap()
    server_addr = osdMap.find_server(osd_id)
    if server_addr is None:
        return
    if op == 'setfattr':
        cmd = 'ssh %(server_addr)s setfattr -n user.cephos.phash.settings -v %(split_factor)s /var/lib/ceph/osd/ceph-%(osd_id)s/current/%(pg_name)s_head' % \
            {"server_addr":server_addr,"split_factor":split_factor,"osd_id":osd_id,"pg_name":pg_name}
    if op == 'getfattr':
        cmd = 'ssh %(server_addr)s getfattr -n user.cephos.phash.settings /var/lib/ceph/osd/ceph-%(osd_id)s/current/%(pg_name)s_head' % \
            {"server_addr":server_addr,"osd_id":osd_id,"pg_name":pg_name}
    print execute_cmd(cmd)

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print "usage: ./split_factor_op.py getfattr"
        print "   or: ./split_factor_op.py setfattr"
        print "\n error: too few arguments"
        sys.exit()

    with open("pgs_detail.txt", 'r') as f:
        pgs = f.readlines()
        for pg in pgs:
            pwds = pg.split(";")
            pg_name = pwds[0]
            osds = pwds[1].replace('[', '').replace(']', '')
            osd_list = osds.split(",")
            for i in range(len(osd_list)):
                rand = random.randint(0, 2)
                op_split_factor(sys.argv[1], int(osd_list[i]), pg_name, split_factors[i][rand])